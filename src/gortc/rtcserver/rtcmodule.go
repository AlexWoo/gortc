// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Server Module

package rtcserver

import (
	"net/http"
	"os"
	"rtclib"
	"strings"

	"github.com/go-ini/ini"
)

type RTCServerConfig struct {
	LogPath       string
	LogLevel      string
	LogRotateSize rtclib.Size_t
	Listen        string
	TlsListen     string
	Cert          string
	Key           string
	Realm         string
	Location      string `default:"/rtc"`
}

type RTCServerModule struct {
	rtcPath   string
	config    *RTCServerConfig
	server    *http.Server
	tlsServer *http.Server
}

var rtcServerModule *RTCServerModule

func NewRTCServerModule() *RTCServerModule {
	rtcServerModule = &RTCServerModule{}

	return rtcServerModule
}

func (m *RTCServerModule) LoadConfig(rtcPath string) bool {
	m.rtcPath = rtcPath
	m.config = new(RTCServerConfig)

	confPath := rtcPath + "/conf/gortc.ini"

	f, err := ini.Load(confPath)
	if err != nil {
		LogError("Load config file %s error: %v", confPath, err)
		return false
	}

	return rtclib.Config(f, "RTCServer", m.config)
}

func (m *RTCServerModule) Init() bool {
	initLog(m.config, m.rtcPath)

	if m.config.Realm == "" {
		LogError("Local Realm not configured")
		return false
	}

	serveMux := &http.ServeMux{}
	serveMux.HandleFunc(m.config.Location, rtcserver)

	if m.config.Listen != "" {
		m.server = &http.Server{Addr: m.config.Listen, Handler: serveMux}
	}

	if m.config.TlsListen != "" {
		if m.config.Cert == "" || m.config.Key == "" {
			LogError("TLS cert(%s) or key(%s) file configured error",
				m.config.Cert, m.config.Key)
			return false
		}

		if !strings.HasPrefix(m.config.Cert, "/") &&
			!strings.HasPrefix(m.config.Cert, "./") {

			m.config.Cert = m.rtcPath + m.config.Cert
		}

		_, err := os.Stat(m.config.Cert)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Cert, err)
			return false
		}

		if !strings.HasPrefix(m.config.Key, "/") &&
			!strings.HasPrefix(m.config.Key, "./") {

			m.config.Key = m.rtcPath + m.config.Key
		}

		_, err = os.Stat(m.config.Key)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Key, err)
			return false
		}

		m.tlsServer = &http.Server{Addr: m.config.TlsListen, Handler: serveMux}
	}

	return true
}

func (m *RTCServerModule) Run() {
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
			err := m.server.ListenAndServe()
			LogError("RTCServer quit, %v", err)
			quit <- true
		}()
	}

	if m.tlsServer != nil {
		LogInfo("RTCServer TLS start ...")
		go func() {
			err := m.tlsServer.ListenAndServeTLS(m.config.Cert, m.config.Key)
			LogError("RTCServer TLS quit, %v", err)
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

func (m *RTCServerModule) Exit() {
	if m.server != nil {
		LogInfo("close RTCServer ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		LogInfo("close RTCServer TLS ...")
		m.tlsServer.Close()
	}
}
