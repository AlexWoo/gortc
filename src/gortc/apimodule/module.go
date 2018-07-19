// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Module

package apimodule

import (
	"net/http"
	"os"

	"github.com/alexwoo/golib"
	"github.com/go-ini/ini"
)

type APIModuleConfig struct {
	LogLevel      string
	LogRotateSize golib.Size
	Listen        string
	TlsListen     string
	Cert          string
	Key           string
}

type APIModule struct {
	rtcpath   string
	config    *APIModuleConfig
	server    *http.Server
	tlsServer *http.Server
}

var module *APIModule

func NewAPIModule(rtcpath string) *APIModule {
	module = &APIModule{
		rtcpath: rtcpath,
	}

	return module
}

func (m *APIModule) LoadConfig() bool {
	m.config = new(APIModuleConfig)

	confPath := m.rtcpath + "/conf/gortc.ini"

	f, err := ini.Load(confPath)
	if err != nil {
		LogError("Load config file %s error: %v", confPath, err)
		return false
	}

	return golib.Config(f, "APIModule", m.config)
}

func (m *APIModule) Init() bool {
	initLog(m.config)

	if !initAPIM() {
		LogError("init API Manager failed")
		return false
	}

	serveMux := &http.ServeMux{}
	serveMux.HandleFunc("/", handler)

	if m.config.Listen != "" {
		m.server = &http.Server{Addr: m.config.Listen, Handler: serveMux}
	}

	if m.config.TlsListen != "" {
		if m.config.Cert == "" || m.config.Key == "" {
			LogError("TLS cert(%s) or key(%s) file configured error",
				m.config.Cert, m.config.Key)
			return false
		}

		m.config.Cert = m.rtcpath + "/certs/" + m.config.Cert

		_, err := os.Stat(m.config.Cert)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Cert, err)
			return false
		}

		m.config.Key = m.rtcpath + "/certs/" + m.config.Key

		_, err = os.Stat(m.config.Key)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Key, err)
			return false
		}

		m.tlsServer = &http.Server{Addr: m.config.TlsListen, Handler: serveMux}
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
			err := m.server.ListenAndServe()
			LogError("APIServer quit, %v", err)
			quit <- true
		}()
	}

	if m.tlsServer != nil {
		LogInfo("APIServer TLS start ...")
		go func() {
			err := m.tlsServer.ListenAndServeTLS(m.config.Cert, m.config.Key)
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
