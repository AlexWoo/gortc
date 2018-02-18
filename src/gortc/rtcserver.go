// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Server Module

package main

import (
	"fmt"
	"net/http"

	"github.com/go-ini/ini"
)

type RTCServerConfig struct {
	Listen    string
	TlsListen string
	Cert      string
	Key       string
	Location  string `default:"/rtc"`
}

type RTCServerModule struct {
	name      string
	config    *RTCServerConfig
	server    *http.Server
	tlsServer *http.Server
}

func NewRTCServerModule() *RTCServerModule {
	m := new(RTCServerModule)
	m.name = "RTCServer"

	return m
}

func (m *RTCServerModule) LoadConfig() bool {
	m.config = new(RTCServerConfig)

	f, err := ini.Load(CONFPATH)
	if err != nil {
		// TODO logErr
		return false
	}

	return Config(f, "RTCServer", m.config)
}

func rtcHandler(w http.ResponseWriter, req *http.Request) {
	//TODO websocket handler
	fmt.Fprintln(w, "RTCServer")
}

func (m *RTCServerModule) Init() bool {
	serveMux := &http.ServeMux{}
	serveMux.HandleFunc(m.config.Location, rtcHandler)

	if m.config.Listen != "" {
		m.server = &http.Server{Addr: m.config.Listen, Handler: serveMux}
	}

	if m.config.TlsListen != "" {
		//TODO tls config check
		m.tlsServer = &http.Server{Addr: m.config.TlsListen, Handler: serveMux}
	}

	//TODO check port conflict with apiserver

	return true
}

func (m *RTCServerModule) Run() {
	if m.server != nil {
		go func() {
			m.server.ListenAndServe()
			//TODO logErr
		}()
	}

	if m.tlsServer != nil {
		go func() {
			m.tlsServer.ListenAndServeTLS(m.config.Cert, m.config.Key)
			//TODO logErr
		}()
	}
}

func (m *RTCServerModule) State() {
}

func (m *RTCServerModule) Exit() {
}
