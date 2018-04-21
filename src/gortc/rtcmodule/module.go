// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Module

package rtcmodule

import (
	"net/http"
	"os"
	"rtclib"
	"strings"

	"github.com/go-ini/ini"
)

type RTCModuleConfig struct {
	LogLevel      string
	LogRotateSize rtclib.Size_t
	Listen        string
	TlsListen     string
	Cert          string
	Key           string
}

type RTCModule struct {
	config    *RTCModuleConfig
	server    *http.Server
	tlsServer *http.Server
	jstack    *rtclib.JSIPStack
}

var module *RTCModule

func process(jsip *rtclib.JSIP) {
	dlg := jsip.DialogueID
	t := rtclib.GetTask(dlg)
	if t != nil {
		t.Process(jsip)
		return
	}

	slpname := "default"

	if len(jsip.Router) != 0 {
		router0 := jsip.Router[0]
		_, _, paras := rtclib.JsipParseUri(router0)

		for _, para := range paras {
			if strings.HasPrefix(para, "type=") {
				ss := strings.SplitN(para, "=", 2)
				if ss[1] != "" {
					slpname = ss[1]
				}
			}
		}
	}

	t = rtclib.NewTask(dlg)
	t.Name = slpname
	slp := getSLP(t)
	if slp == nil {
		rtclib.SendJSIPRes(jsip, 404)
		t.DelTask()
		return
	}

	t.SLP = slp
	t.Process(jsip)
}

func NewRTCModule() *RTCModule {
	module = &RTCModule{}

	return module
}

func (m *RTCModule) LoadConfig() bool {
	m.config = new(RTCModuleConfig)

	confPath := rtclib.RTCPATH + "/conf/gortc.ini"

	f, err := ini.Load(confPath)
	if err != nil {
		LogError("Load config file %s error: %v", confPath, err)
		return false
	}

	return rtclib.Config(f, "RTCModule", m.config)
}

func (m *RTCModule) Init() bool {
	initLog(m.config)

	if !initSLPM() {
		LogError("SLP Manager init error")
		return false
	}

	m.jstack = rtclib.InitJSIPStack(process, log)
	if m.jstack == nil {
		LogError("JSIP Stack init error")
		return false
	}

	serveMux := &http.ServeMux{}
	serveMux.HandleFunc(m.jstack.Location(), m.jstack.RTCServer)

	if m.config.Listen != "" {
		m.server = &http.Server{Addr: m.config.Listen, Handler: serveMux}
	}

	if m.config.TlsListen != "" {
		if m.config.Cert == "" || m.config.Key == "" {
			LogError("TLS cert(%s) or key(%s) file configured error",
				m.config.Cert, m.config.Key)
			return false
		}

		m.config.Cert = rtclib.RTCPATH + "/certs/" + m.config.Cert

		_, err := os.Stat(m.config.Cert)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Cert, err)
			return false
		}

		m.config.Key = rtclib.RTCPATH + "/certs/" + m.config.Key

		_, err = os.Stat(m.config.Key)
		if err != nil {
			LogError("TLS cert(%s) error: %v", m.config.Key, err)
			return false
		}

		m.tlsServer = &http.Server{Addr: m.config.TlsListen, Handler: serveMux}
	}

	return true
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
