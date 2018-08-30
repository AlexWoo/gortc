// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// rtcserver Module

package main

import (
	"fmt"
	"net/http"
	"os"
	"rtclib"
	"strconv"
	"time"

	"github.com/alexwoo/golib"
	"github.com/gorilla/websocket"
)

// Normal Config
type rtcConfig struct {
	Listen    string
	TlsListen string
	Cert      string
	Key       string
}

// Dynamic Config which can be reload
type rtcDConfig struct {
	LogFile             string        `default:"logs/rtc.log"`
	LogLevel            string        `default:"info"`
	ClientHeaderTimeout time.Duration `default:"10s"`
	Keepalived          time.Duration `default:"60s"`
	AccessFile          string        `default:"logs/access.log"`
}

type rtcServer struct {
	config    *rtcConfig
	dconfig   *rtcDConfig
	log       *golib.Log
	logLevel  int
	server    *golib.HTTPServer
	tlsServer *golib.HTTPServer
	nServers  uint

	jsipC  chan *rtclib.JSIP
	taskQ  chan *rtclib.Task
	jstack *rtclib.JSIPStack
}

var rtcs *rtcServer

func rtcServerInstance() *rtcServer {
	if rtcs != nil {
		return rtcs
	}

	rtcs = &rtcServer{}

	return rtcs
}

func (m *rtcServer) loadDConfig() error {
	confPath := rtclib.FullPath("conf/gortc.ini")

	config := &rtcDConfig{}
	err := golib.ConfigFile(confPath, "RTCModule", config)
	if err != nil {
		return fmt.Errorf("Parse dconfig %s Failed, %s", confPath, err)
	}
	m.dconfig = config

	return nil
}

func (m *rtcServer) loadConfig() error {
	confPath := rtclib.FullPath("conf/gortc.ini")

	config := &rtcConfig{}
	err := golib.ConfigFile(confPath, "RTCModule", config)
	if err != nil {
		return fmt.Errorf("Parse config %s Failed, %s", confPath, err)
	}
	m.config = config

	return nil
}

func (m *rtcServer) initLog() error {
	logPath := rtclib.FullPath(m.dconfig.LogFile)
	m.logLevel = golib.LoglvEnum.ConfEnum(m.dconfig.LogLevel, golib.LOGINFO)
	m.log = golib.NewLog(logPath)

	return nil
}

// rtc handler

func (m *rtcServer) wsCheckOrigin(r *http.Request) bool {
	//Access Control from here
	return true
}

func (m *rtcServer) handler(w http.ResponseWriter, req *http.Request) {
	userid := req.URL.Query().Get("userid")
	if userid == "" {
		m.LogError("Miss userid")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  64 * 1024,
		WriteBufferSize: 64 * 1024,
		CheckOrigin:     m.wsCheckOrigin,
	}

	c, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		m.LogError("Create Websocket server failed, %v", err)
		return
	}

	conn := golib.NewWSServer(userid, c, m.jstack.Qsize(), rtclib.RecvMsg,
		m.log, m.logLevel)

	conn.Accept()
}

func (m *rtcServer) processMsg(jsip *rtclib.JSIP) {
	dlg := jsip.DialogueID
	t := rtclib.GetTask(dlg)
	if t != nil {
		t.OnMsg(jsip)
		return
	}

	if jsip.Code > 0 {
		m.LogError("Receive %s but SLP if finished", jsip.Name())
		return
	}

	if jsip.Type != rtclib.INVITE && jsip.Type != rtclib.REGISTER &&
		jsip.Type != rtclib.OPTIONS && jsip.Type != rtclib.MESSAGE &&
		jsip.Type != rtclib.SUBSCRIBE {

		m.LogError("Receive %s but SLP if finished", jsip.Name())
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

	t = rtclib.NewTask(dlg, m.taskQ, m.log, m.logLevel)
	t.Name = slpname
	sm.getSLP(t, SLPPROCESS)
	if t.SLP == nil {
		rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 404))
		t.DelTask()
		return
	}

	t.OnMsg(jsip)
}

func (m *rtcServer) process() {
	for {
		select {
		case jsip := <-m.jsipC:
			m.processMsg(jsip)
		case task := <-m.taskQ:
			task.DelTask()
		}
	}
}

// for module interface

func (m *rtcServer) PreInit() error {
	if err := m.loadConfig(); err != nil {
		return err
	}

	if err := m.loadDConfig(); err != nil {
		return err
	}

	if err := m.initLog(); err != nil {
		return err
	}

	return nil
}

func (m *rtcServer) Init() error {
	m.jsipC = make(chan *rtclib.JSIP, 4096)
	m.taskQ = make(chan *rtclib.Task, 1024)

	m.jstack = rtclib.InitJSIPStack(m.jsipC, m.log, m.logLevel)

	if m.jstack == nil {
		return fmt.Errorf("JSIP Stack init error")
	}

	accessFile := rtclib.FullPath(m.dconfig.AccessFile)

	if m.config.Listen != "" {
		s, err := golib.NewHTTPServer(m.config.Listen, "", "", "/",
			m.dconfig.ClientHeaderTimeout, m.dconfig.Keepalived, m.log,
			m.handler, accessFile)
		if err != nil {
			return fmt.Errorf("New API Server error: %s", err)
		}

		m.server = s

		m.nServers++
	}

	if m.config.TlsListen != "" {
		if m.config.Cert == "" || m.config.Key == "" {
			return fmt.Errorf("TLS cert(%s) or key(%s) file configured error",
				m.config.Cert, m.config.Key)
		}

		m.config.Cert = rtclib.FullPath("certs/" + m.config.Cert)
		m.config.Key = rtclib.FullPath("certs/" + m.config.Key)

		s, err := golib.NewHTTPServer(m.config.Listen, m.config.Cert,
			m.config.Key, "/", m.dconfig.ClientHeaderTimeout,
			m.dconfig.Keepalived, m.log, m.handler, accessFile)
		if err != nil {
			return fmt.Errorf("New API TLSServer error: %s", err)
		}

		m.tlsServer = s

		m.nServers++
	}

	return nil
}

func (m *rtcServer) PreMainloop() error {
	return nil
}

func (m *rtcServer) Mainloop() {
	quit := make(chan bool)

	if m.server != nil {
		m.LogInfo("rtc server start ...")
		go func() {
			err := m.server.Start()
			if err != nil {
				m.LogError("rtc server quit, %s", err)
			}
			quit <- true
		}()
	}

	if m.tlsServer != nil {
		m.LogInfo("rtc server start ...")
		go func() {
			err := m.tlsServer.Start()
			if err != nil {
				m.LogError("rtc tlsserver quit, %s", err)
			}
			quit <- true
		}()
	}

	go m.process()

	for {
		<-quit
		m.nServers--

		if m.nServers == 0 {
			break
		}
	}
}

func (m *rtcServer) Reload() error {
	if err := m.loadDConfig(); err != nil {
		return err
	}

	if err := m.initLog(); err != nil {
		return err
	}

	return nil
}

func (m *rtcServer) Reopen() error {
	if err := m.initLog(); err != nil {
		return err
	}

	return nil
}

func (m *rtcServer) Exit() {
	if m.server != nil {
		m.LogInfo("closing api server ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		m.LogInfo("closing api tlsserver ...")
		m.tlsServer.Close()
	}
}

// for log ctx

func (m *rtcServer) Prefix() string {
	return "[rtc] " + strconv.Itoa(os.Getpid())
}

func (m *rtcServer) Suffix() string {
	return ""
}

func (m *rtcServer) LogLevel() int {
	return m.logLevel
}

// for log ctx

func (m *rtcServer) LogDebug(format string, v ...interface{}) {
	m.log.LogDebug(m, format, v...)
}

func (m *rtcServer) LogInfo(format string, v ...interface{}) {
	m.log.LogInfo(m, format, v...)
}

func (m *rtcServer) LogError(format string, v ...interface{}) {
	m.log.LogError(m, format, v...)
}

func (m *rtcServer) LogFatal(format string, v ...interface{}) {
	m.log.LogFatal(m, format, v...)
}
