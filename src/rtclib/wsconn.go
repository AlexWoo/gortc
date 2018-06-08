// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Websocket Connection

package rtclib

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSConn struct {
	conn      *websocket.Conn
	name      string
	url       string
	uatype    int
	timeout   time.Duration
	sendq     chan []byte
	recvq     chan []byte
	reconnect chan bool
	quit      chan bool
	buf       []byte
	handler   func(c Conn, data []byte)
}

var (
	wsconns     = make(map[string]*WSConn)
	wsconnsLock sync.Mutex
)

func (c *WSConn) connect() bool {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.timeout

	for {
		conn, _, err := dialer.Dial(c.url, nil)
		if err == nil { // connect successd
			go c.read()
			go c.loop()
			c.conn = conn

			return true
		}

		fmt.Println(time.Now(), "------ connect err: ", err)

		if websocket.IsCloseError(err) {
			e := err.(*websocket.CloseError)

			switch e.Code {
			case websocket.CloseTLSHandshake:
				if strings.HasPrefix(c.url, "ws://") {
					strings.Replace(c.url, "ws://", "wss://", 1)
					continue
				}
			}

			fmt.Println(e.Code)
		} else {
			e, ok := err.(net.Error)
			if ok {
				if e.Timeout() {
					continue
				}
			}
		}

		time.Sleep(c.timeout)
	}
}

func (c *WSConn) read() {
	for {
		conn := c.conn
		_, data, err := conn.ReadMessage()
		if err == nil {
			c.recvq <- data
			continue
		}

		fmt.Println(time.Now(), "------ read err: ", err)

		c.recvq <- []byte{}

		return
	}
}

func (c *WSConn) write(data []byte) bool {
	err := c.conn.WriteMessage(websocket.TextMessage, data)
	if err == nil {
		c.buf = []byte{}
		return true
	}

	fmt.Println(time.Now(), "------ write err: ", err)

	c.buf = data

	if websocket.IsCloseError(err) {
		e := err.(*websocket.CloseError)

		switch e.Code {
		case websocket.CloseMessageTooBig:
			c.buf = []byte{}
		}
	}

	return false
}

func (c *WSConn) loop() {
	var data []byte

	if len(c.buf) != 0 {
		data = c.buf
		if !c.write(data) {
			return
		}
	}

	for {
		select {
		case data = <-c.recvq:
			if len(data) == 0 {
				c.reconnect <- true
				return
			}
			c.handler(c, data)
		case data = <-c.sendq:
			if !c.write(data) {
				c.reconnect <- true
				return
			}
		}
	}
}

// New a websocket Conn instance
// 	name   : connection name
// 	url    : client connection need url to connect,
//     must be ws[s]://domain/location?paras
// 	uatype : if connection is client, set to UAC, otherwise, set to UAS
// 	qsize  : connection send queue size
// 	handler: handler to call when receive data from websocket connection
func NewWSConn(name string, url string, uatype int, timeout time.Duration,
	qsize uint64, handler func(c Conn, data []byte)) *WSConn {

	if uatype == UAC && !(strings.HasPrefix(url, "ws://") ||
		strings.HasPrefix(url, "wss://")) {

		return nil
	}

	wsconnsLock.Lock()
	defer wsconnsLock.Unlock()

	conn := wsconns[name]
	if conn != nil {
		return conn
	}

	conn = &WSConn{
		name:      name,
		url:       url,
		uatype:    uatype,
		timeout:   timeout,
		sendq:     make(chan []byte, qsize),
		recvq:     make(chan []byte, qsize),
		reconnect: make(chan bool),
		quit:      make(chan bool),
		handler:   handler,
	}
	wsconns[name] = conn

	return conn
}

// Dial to remote websocket server
func (c *WSConn) Dial() {
	go func() {
		c.connect()

		for {
			select {
			case <-c.reconnect:
				c.connect()
			case <-c.quit:
				c.conn.Close()
				return
			}
		}
	}()
}

// Accept remote websocket connection
//	conn: websocket conn
func (c *WSConn) Accept(conn interface{}) {
	c.conn = conn.(*websocket.Conn)

	go c.read()
	go c.loop()

	for {
		select {
		case <-c.reconnect:
			return
		case <-c.quit:
			c.conn.Close()
			return
		}
	}
}

// Send data through websocket connection
// 	data: data need to send, now only support text message
func (c *WSConn) Send(data []byte) {
	c.sendq <- data
}

// Close websocket connection
func (c *WSConn) Close() {
	wsconnsLock.Lock()
	delete(wsconns, c.name)
	wsconnsLock.Unlock()

	c.quit <- true
}
