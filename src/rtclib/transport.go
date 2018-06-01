// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Transport layer

package rtclib

import (
	"net/http"
	"strings"
	"sync"

	"github.com/go-ini/ini"
	"github.com/gorilla/websocket"
)

type JSIPConfig struct {
	Location string `default:"rtc"`
	Realm    string
}

type JSIPConn struct {
	conn   *websocket.Conn
	uatype int
	sendq  chan interface{}
}

type JSIPStack struct {
	config     *JSIPConfig
	jsipHandle func(jsip *JSIP)
	log        *Log
	conns      sync.Map //map[string]*JSIPConn
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  64 * 1024,
	WriteBufferSize: 64 * 1024,
	CheckOrigin:     wsCheckOrigin,
}

var jstack *JSIPStack

func InitJSIPStack(h func(jsip *JSIP), log *Log) *JSIPStack {
	jstack = &JSIPStack{
		jsipHandle: h,
		log:        log,
	}

	if !jstack.LoadConfig() {
		return nil
	}

	if jstack.config.Realm == "" {
		jstack.log.LogError("JSIPStack Realm not configured")
		return nil
	}

	return jstack
}

func (jconn *JSIPConn) write() {
	defer jconn.conn.Close()

	for {
		select {
		case msg, ok := <-jconn.sendq:
			if !ok {
				jstack.log.LogError("RTC write message error")
				return
			}

			jconn.conn.WriteJSON(msg)
		}
	}
}

func (jconn *JSIPConn) read() {
	defer jconn.conn.Close()

	for {
		_, data, err := jconn.conn.ReadMessage()
		if err != nil {
			jstack.log.LogError("RTC read message error, %v", err)
			return
		}

		RecvJsonSIPMsg(jconn, data)
	}
}

func wsCheckOrigin(r *http.Request) bool {
	//Access Control from here
	return true
}

func (stack *JSIPStack) RTCServer(w http.ResponseWriter, req *http.Request) {
	userid := req.URL.Query().Get("userid")
	if userid == "" {
		jstack.log.LogError("Miss userid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		jstack.log.LogError("Create Websocket server failed, %v", err)
		return
	}

	jconn := &JSIPConn{
		conn:   conn,
		uatype: UAS,
		sendq:  make(chan interface{}, 1024),
	}

	vv, ok := stack.conns.LoadOrStore(userid, jconn)
	if ok {
		jconn = vv.(*JSIPConn)
		if jconn.uatype == UAC {
			jstack.log.LogError("%s should be UAC, refused connection", userid)
			return
		}
		jconn.conn = conn
	}

	go jconn.write()
	jconn.read()
}

func parseUri(uri string) (string, string) {
	var userWithHost, hostWithPort string

	ss := strings.Split(uri, ";")
	uri = ss[0]

	ss = strings.Split(uri, "@")
	if len(ss) == 1 {
		hostWithPort = ss[0]
	} else {
		hostWithPort = ss[1]
	}

	ss = strings.Split(uri, ":")
	userWithHost = ss[0]

	return userWithHost, hostWithPort
}

func (stack *JSIPStack) RTCClient(uri string) *JSIPConn {
	userWithHost, hostWithPort := parseUri(uri)
	if userWithHost == "" {
		return nil
	}

	jconn := &JSIPConn{
		uatype: UAC,
		sendq:  make(chan interface{}, 1024),
	}

	vv, ok := stack.conns.LoadOrStore(userWithHost, jconn)
	if ok {
		jconn = vv.(*JSIPConn)
		return jconn
	}

	dialer := websocket.DefaultDialer
	url := "ws://" + hostWithPort + stack.config.Location +
		"?userid=" + stack.config.Realm
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		jstack.log.LogError("Connect to %s failed", url)
		return nil
	}
	jconn.conn = conn

	go jconn.read()
	go jconn.write()

	return jconn
}

func (stack *JSIPStack) LoadConfig() bool {
	stack.config = new(JSIPConfig)

	confPath := RTCPATH + "/conf/gortc.ini"

	f, err := ini.Load(confPath)
	if err != nil {
		jstack.log.LogError("Load config file %s error: %v", confPath, err)
		return false
	}

	return Config(f, "JSIPStack", stack.config)
}

func (stack *JSIPStack) Location() string {
	return stack.config.Location
}

func (stack *JSIPStack) Realm() string {
	return stack.config.Realm
}
