// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Module

package rtcmodule

import (
	"fmt"
	"net/http"
	"rtclib"
	"time"

	"github.com/alexwoo/golib"
	"github.com/gorilla/websocket"
)

type RTCModuleConfig struct {
	LogFile             string
	LogLevel            string `default:"info"`
	Listen              string
	TlsListen           string
	Cert                string
	Key                 string
	ClientHeaderTimeout time.Duration `default:"10s"`
	Keepalived          time.Duration `default:"60s"`
	AccessFile          string        `default:"logs/access.log"`
}

type RTCModule struct {
	rtcpath string

	config    *RTCModuleConfig
	server    *golib.HTTPServer
	tlsServer *golib.HTTPServer

	jsipC  chan *rtclib.JSIP
	taskQ  chan *rtclib.Task
	jstack *rtclib.JSIPStack
}

var module *RTCModule

func NewRTCModule(rtcpath string) *RTCModule {
	module = &RTCModule{
		rtcpath: rtcpath,
	}

	return module
}

func (m *RTCModule) LoadConfig() bool {
	m.config = new(RTCModuleConfig)
	confPath := rtclib.FullPath("conf/gortc.ini")

	err := golib.ConfigFile(confPath, "RTCModule", m.config)
	if err != nil {
		fmt.Printf("Parse config %s error: %v\n", confPath, err)
		return false
	}

	return true
}

func wsCheckOrigin(r *http.Request) bool {
	//Access Control from here
	return true
}

func (m *RTCModule) handler(w http.ResponseWriter, req *http.Request) {
	userid := req.URL.Query().Get("userid")
	if userid == "" {
		LogError("Miss userid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  64 * 1024,
		WriteBufferSize: 64 * 1024,
		CheckOrigin:     wsCheckOrigin,
	}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		LogError("Create Websocket server failed, %v", err)
		return
	}

	conn := golib.NewWSServer(userid, c, m.jstack.Qsize(), rtclib.RecvMsg,
		rtclogCtx.log, rtclogCtx.logLevel)

	conn.Accept()
}

func (m *RTCModule) Init(log *golib.Log) bool {
	initLog(m.config, log)

	if !initSLPM() {
		LogError("SLP Manager init error")
		return false
	}

	m.jsipC = make(chan *rtclib.JSIP, 4096)
	m.taskQ = make(chan *rtclib.Task, 1024)

	m.jstack = rtclib.InitJSIPStack(m.jsipC, rtclogCtx.log, rtclogCtx.logLevel,
		m.rtcpath)
	if m.jstack == nil {
		LogError("JSIP Stack init error")
		return false
	}

	m.config.AccessFile = rtclib.FullPath(m.config.AccessFile)

	if m.config.Listen != "" {
		s, err := golib.NewHTTPServer(m.config.Listen, "", "",
			m.jstack.Location(), m.config.ClientHeaderTimeout,
			m.config.Keepalived, rtclogCtx.log, m.handler, m.config.AccessFile)
		if err != nil {
			LogError("New RTC Server error: %s", err)
			return false
		}

		m.server = s
	}

	if m.config.TlsListen != "" {
		if m.config.Cert == "" || m.config.Key == "" {
			LogError("TLS cert(%s) or key(%s) file configured error",
				m.config.Cert, m.config.Key)
			return false
		}

		m.config.Cert = rtclib.FullPath("certs/" + m.config.Cert)
		m.config.Key = rtclib.FullPath("certs/" + m.config.Key)

		s, err := golib.NewHTTPServer(m.config.Listen, m.config.Cert,
			m.config.Key, m.jstack.Location(), m.config.ClientHeaderTimeout,
			m.config.Keepalived, rtclogCtx.log, m.handler, m.config.AccessFile)
		if err != nil {
			LogError("New RTC Server error: %s", err)
			return false
		}

		m.tlsServer = s
	}

	return true
}

func (m *RTCModule) processMsg(jsip *rtclib.JSIP) {
	dlg := jsip.DialogueID
	t := rtclib.GetTask(dlg)
	if t != nil {
		t.OnMsg(jsip)
		return
	}

	if jsip.Code > 0 {
		LogError("Receive %s but SLP if finished", jsip.Name())
		return
	}

	if jsip.Type != rtclib.INVITE && jsip.Type != rtclib.REGISTER &&
		jsip.Type != rtclib.OPTIONS && jsip.Type != rtclib.MESSAGE &&
		jsip.Type != rtclib.SUBSCRIBE {

		LogError("Receive %s but SLP if finished", jsip.Name())
		return
	}

	slpname := "default"

	if len(jsip.Router) != 0 {
		jsipUri, err := rtclib.ParseJSIPUri(jsip.Router[0])
		if err != nil {
			rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 400))
			return
		}

		relid, ok := jsipUri.Paras["relid"].(string)
		if ok {
			t := rtclib.GetTask(relid)
			if t == nil {
				rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 400))
				return
			}

			t.OnMsg(jsip)
			return
		}

		name, ok := jsipUri.Paras["type"].(string)
		if ok && name != "" {
			slpname = name
		}
	}

	t = rtclib.NewTask(dlg, m.taskQ, rtclogCtx.log, rtclogCtx.logLevel)
	t.Name = slpname
	getSLP(t, SLPPROCESS)
	if t.SLP == nil {
		rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 404))
		t.DelTask()
		return
	}

	t.OnMsg(jsip)
}

func (m *RTCModule) process() {
	for {
		select {
		case jsip := <-m.jsipC:
			m.processMsg(jsip)
		case task := <-m.taskQ:
			task.DelTask()
		}
	}
}

func (m *RTCModule) Run() {
	wait := 0
	if m.server != nil {
		wait++
	}
	if m.tlsServer != nil {
		wait++
	}
	quit := make(chan bool, wait)

	if m.server != nil {
		LogInfo("RTCServer start ...")
		go func() {
			//TODO retry
			err := m.server.Start()
			LogError("RTCServer quit, %v", err)
			quit <- true
		}()
	}

	if m.tlsServer != nil {
		LogInfo("RTCServer TLS start ...")
		go func() {
			err := m.tlsServer.Start()
			LogError("RTCServer TLS quit, %v", err)
			quit <- true
		}()
	}

	go m.process()

	for {
		<-quit
		wait--

		if wait == 0 {
			break
		}
	}
}

func (m *RTCModule) Exit() {
	if m.server != nil {
		LogInfo("close RTCServer ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		LogInfo("close RTCServer TLS ...")
		m.tlsServer.Close()
	}
}
