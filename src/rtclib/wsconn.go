// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Websocket Connection

package rtclib

import (
	"fmt"
	"net"
	"strings"
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
	reconnect chan *websocket.Conn
	quit      chan bool
	buf       []byte
	handler   func(c Conn, data []byte)
}

func (c *WSConn) connect() bool {
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = c.timeout

	for {
		conn, _, err := dialer.Dial(c.url, nil)
		if err == nil { // connect successd
			go c.read()
			go c.write()
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
			c.handler(c, data)
			continue
		}

		fmt.Println(time.Now(), "------ read err: ", err)

		c.reconnect <- conn

		return
	}
}

func (c *WSConn) write() {
	var data []byte

	for {
		if len(c.buf) != 0 {
			data = c.buf
		} else {
			data = <-c.sendq
		}

		conn := c.conn
		err := conn.WriteMessage(websocket.TextMessage, data)
		if err == nil {
			c.buf = []byte{}
			continue
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

		c.reconnect <- conn

		return
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
	qsize uint32, handler func(c Conn, data []byte)) *WSConn {

	if uatype == UAC && url == "" {
		return nil
	}

	conn := &WSConn{
		name:      name,
		url:       url,
		uatype:    uatype,
		timeout:   timeout,
		sendq:     make(chan []byte, qsize),
		reconnect: make(chan *websocket.Conn, 2),
		quit:      make(chan bool),
		handler:   handler,
	}

	return conn
}

// Set websocket connection to WSConn
// 	conn: websocket.Conn
func (c *WSConn) SetConn(conn *websocket.Conn) {
	c.conn = conn
}

// Dial to remote websocket server
func (c *WSConn) Dial() {
	if !c.connect() {
		return
	}

	go func() {
		for {
			select {
			case conn := <-c.reconnect:
				if conn != c.conn {
					continue
				}

				if !c.connect() { // reconnect
					return
				}
			case <-c.quit:
				c.conn.Close()
				return
			}
		}
	}()
}

// Accept remote websocket connection
func (c *WSConn) Accept() {
	go c.read()
	go c.write()

	for {
		select {
		case conn := <-c.reconnect:
			if conn != c.conn {
				continue
			}

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
	c.quit <- true
}
