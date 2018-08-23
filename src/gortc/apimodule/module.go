// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Module

package apimodule

import (
	"fmt"
	"rtclib"
	"time"

	"github.com/alexwoo/golib"
)

type APIModuleConfig struct {
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

type APIModule struct {
	config    *APIModuleConfig
	server    *golib.HTTPServer
	tlsServer *golib.HTTPServer
}

var module *APIModule

func NewAPIModule(rtcpath string) *APIModule {
	module = &APIModule{}

	return module
}

func (m *APIModule) LoadConfig() bool {
	m.config = new(APIModuleConfig)
	confPath := rtclib.FullPath("conf/gortc.ini")

	err := golib.ConfigFile(confPath, "APIModule", m.config)
	if err != nil {
		fmt.Printf("Parse config %s error: %v\n", confPath, err)
		return false
	}

	return true
}

func (m *APIModule) Init(log *golib.Log) bool {
	initLog(m.config, log)

	if !initAPIM() {
		LogError("init API Manager failed")
		return false
	}

	m.config.AccessFile = rtclib.FullPath(m.config.AccessFile)

	if m.config.Listen != "" {
		s, err := golib.NewHTTPServer(m.config.Listen, "", "", "/",
			m.config.ClientHeaderTimeout, m.config.Keepalived, apilogCtx.log,
			handler, m.config.AccessFile)
		if err != nil {
			LogError("New API Server error: %s", err)
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
			m.config.Key, "/", m.config.ClientHeaderTimeout,
			m.config.Keepalived, apilogCtx.log, handler, m.config.AccessFile)
		if err != nil {
			LogError("New API TLSServer error: %s", err)
			return false
		}

		m.tlsServer = s
	}

	return true
}

func (m *APIModule) Run() {
	wait := 0
	if m.server != nil {
		wait++
	}
	if m.tlsServer != nil {
		wait++
	}
	quit := make(chan bool, wait)

	if m.server != nil {
		LogInfo("APIServer start ...")
		go func() {
			// TODO retry
			err := m.server.Start()
			LogError("APIServer quit, %v", err)
			quit <- true
		}()
	}

	if m.tlsServer != nil {
		LogInfo("APIServer TLS start ...")
		go func() {
			err := m.tlsServer.Start()
			LogError("APIServer TLS quit, %v", err)
			quit <- true
		}()
	}

	for {
		<-quit
		wait--

		if wait == 0 {
			break
		}
	}
}

func (m *APIModule) Exit() {
	if m.server != nil {
		LogInfo("close APIServer ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		LogInfo("close APIServer TLS ...")
		m.tlsServer.Close()
	}
}
