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
	Listen    string `default:":2539"`
	TlsListen string
	Cert      string
	Key       string
}

type APIServerModule struct {
	name   string
	quit   chan bool
	config *APIServerConfig
}

func NewAPIServerModule() *APIServerModule {
	m := new(APIServerModule)
	m.name = "API server"
	m.quit = make(chan bool)

	return m
}

func (m *APIServerModule) LoadConfig() bool {
	m.config = new(APIServerConfig)

	f, err := ini.Load(CONFPATH)
	if err != nil {
		// TODO logErr
		return false
	}

	return Config(f, "APIServer", m.config)
}

func apiHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

func (m *APIServerModule) Init() bool {
	//TODO tls config check
	//TODO api server route
	http.HandleFunc("/", apiHandler)

	return true
}

func (m *APIServerModule) Run() {
	if m.config.Listen != "" {
		go func() {
			err := http.ListenAndServe(m.config.Listen, nil)
			//TODO logErr
			fmt.Println(err)
		}()
	}

	if m.config.TlsListen != "" {
		go func() {
			http.ListenAndServeTLS(m.config.TlsListen, m.config.Cert,
				m.config.Key, nil)
			//TODO logErr
		}()
	}
}

func (m *APIServerModule) State() {
}

func (m *APIServerModule) Exit() {
}
