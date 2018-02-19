// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Server Module

package main

import (
	"fmt"
	"net/http"

	"github.com/go-ini/ini"
)

type APIServerConfig struct {
	Listen    string
	TlsListen string
	Cert      string
	Key       string
}

type APIServerModule struct {
	name      string
	config    *APIServerConfig
	server    *http.Server
	tlsServer *http.Server
}

func NewAPIServerModule() *APIServerModule {
	m := new(APIServerModule)
	m.name = "APIServer"

	return m
}

func (m *APIServerModule) LoadConfig() bool {
	m.config = new(APIServerConfig)

	f, err := ini.Load(CONFPATH)
	if err != nil {
		LogError("Load config file %s error: %v", CONFPATH, err)
		return false
	}

	return Config(f, "APIServer", m.config)
}

func apiHandler(w http.ResponseWriter, req *http.Request) {
	//TODO api server route
	fmt.Fprintln(w, "APIServer")
}

func (m *APIServerModule) Init() bool {
	serveMux := &http.ServeMux{}
	serveMux.HandleFunc("/", apiHandler)

	if m.config.Listen != "" {
		m.server = &http.Server{Addr: m.config.Listen, Handler: serveMux}
	}

	if m.config.TlsListen != "" {
		//TOOD check cert and key
		m.tlsServer = &http.Server{Addr: m.config.TlsListen, Handler: serveMux}
	}

	return true
}

func (m *APIServerModule) Run() {
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

func (m *APIServerModule) State() {
}

func (m *APIServerModule) Exit() {
}
