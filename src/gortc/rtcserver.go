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
	Location  string `default:"/rtc"`
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
	Qsize               uint64        `default:"1024"`
}

type rtcServer struct {
	config    *rtcConfig
	dconfig   *rtcDConfig
	log       *golib.Log
	logLevel  int
	server    *golib.HTTPServer
	tlsServer *golib.HTTPServer
	nServers  uint

	taskQ chan *rtclib.Task
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

	conn := golib.NewWSServer(userid, c, m.dconfig.Qsize, rtclib.RecvMsg,
		m.log, m.logLevel)

	conn.Accept()
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

	rtclib.JStackInstance().SetLog(m.log, m.logLevel)

	return nil
}

func (m *rtcServer) Init() error {
	m.taskQ = make(chan *rtclib.Task, 1024)

	accessFile := rtclib.FullPath(m.dconfig.AccessFile)

	if m.config.Listen != "" {
		s, err := golib.NewHTTPServer(m.config.Listen, "", "",
			m.config.Location, m.dconfig.ClientHeaderTimeout,
			m.dconfig.Keepalived, m.log, m.handler, accessFile)
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

		s, err := golib.NewHTTPServer(m.config.TlsListen, m.config.Cert,
			m.config.Key, m.config.Location, m.dconfig.ClientHeaderTimeout,
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
			retry := 0
			for {
				err := m.server.Start()
				if err != nil {
					retry++
					m.LogError("rtc server start failure %d, %s", retry, err)

					if retry >= 10 {
						m.LogError("rtc server start failure, system exit")
						os.Exit(1)
					}

					time.Sleep(500 * time.Millisecond)
				} else {
					m.LogError("rtc server close")
					quit <- true
					break
				}
			}
		}()
	}

	if m.tlsServer != nil {
		m.LogInfo("rtc tlsserver start ...")
		go func() {
			retry := 0
			for {
				err := m.tlsServer.Start()
				if err != nil {
					retry++
					m.LogError("rtc tlsserver start failure %d, %s", retry, err)

					if retry >= 10 {
						m.LogError("rtc tlsserver start failure, system exit")
						os.Exit(1)
					}

					time.Sleep(500 * time.Millisecond)
				} else {
					m.LogError("rtc tlsserver close")
					quit <- true
					break
				}
			}
		}()
	}

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
		m.LogInfo("closing rtc server ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		m.LogInfo("closing rtc tlsserver ...")
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
