// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Server Module

package apiserver

import (
	"net/http"
	"os"
	"regexp"
	"rtclib"
	"strings"

	"github.com/go-ini/ini"
)

type APIServerConfig struct {
	LogPath       string
	LogLevel      string
	LogRotateSize rtclib.Size_t
	Listen        string
	TlsListen     string
	Cert          string
	Key           string
}

type APIServerModule struct {
	rtcPath   string
	config    *APIServerConfig
	server    *http.Server
	tlsServer *http.Server
}

var apiserver *APIServerModule

func NewAPIServerModule() *APIServerModule {
	apiserver := &APIServerModule{}

	return apiserver
}

func (m *APIServerModule) LoadConfig(rtcPath string) bool {
	m.rtcPath = rtcPath
	m.config = new(APIServerConfig)

	confPath := rtcPath + "/conf/gortc.ini"

	f, err := ini.Load(confPath)
	if err != nil {
		LogError("Load config file %s error: %v", confPath, err)
		return false
	}

	return rtclib.Config(f, "APIServer", m.config)
}

func parseUri(uri string) (bool, string, string, string) {
	reg := regexp.MustCompile(`^/(\w+)/(v\d+)/(.+)`)

	match := reg.FindStringSubmatch(uri)
	if len(match) == 0 {
		return false, "", "", ""
	}

	return true, match[1], match[2], match[3]
}

func callAPI(req *http.Request, apiname string, version string,
	paras string) (int, *map[string]string, interface{}, *map[int]RespCode) {

	return 0, nil, nil, nil
}

func handler(w http.ResponseWriter, req *http.Request) {
	ok, apiname, version, paras := parseUri(req.RequestURI)
	if !ok {
		NewResponse(1, nil, nil, nil).SendResp(w)
		return
	}

	NewResponse(callAPI(req, apiname, version, paras)).SendResp(w)
}

func (m *APIServerModule) Init() bool {
	initLog(m.config, m.rtcPath)

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

func (m *APIServerModule) Run() {
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

func (m *APIServerModule) Exit() {
	if m.server != nil {
		LogInfo("close APIServer ...")
		m.server.Close()
	}

	if m.tlsServer != nil {
		LogInfo("close APIServer TLS ...")
		m.tlsServer.Close()
	}
}
