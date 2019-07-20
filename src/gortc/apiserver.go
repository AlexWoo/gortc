// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// apiserver Module

package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"rtclib"
	"strconv"
	"strings"
	"time"

	"github.com/alexwoo/golib"
)

// Normal Config
type apiConfig struct {
	Listen    string
	TlsListen string
	Cert      string
	Key       string
}

// Dynamic Config which can be reload
type apiDConfig struct {
	LogFile             string        `default:"logs/rtc.log"`
	LogLevel            string        `default:"info"`
	ClientHeaderTimeout time.Duration `default:"10s"`
	Keepalived          time.Duration `default:"60s"`
	AccessFile          string        `default:"logs/access.log"`
}

type apiServer struct {
	config    *apiConfig
	dconfig   *apiDConfig
	log       *golib.Log
	logLevel  int
	server    *golib.HTTPServer
	tlsServer *golib.HTTPServer
	nServers  uint
}

var apis *apiServer

func apiServerInstance() *apiServer {
	if apis != nil {
		return apis
	}

	apis = &apiServer{}

	return apis
}

func fullAddr(addr string) string {
	split := strings.Split(addr, ":")

	if split[0] == "" {
		return "127.0.0.1" + addr
	}

	return addr
}

func (m *apiServer) writeAPIFile() {
	apifile := rtclib.FullPath("bin/.api")
	f, err := os.OpenFile(apifile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Println("Write pid file failed", err)
		os.Exit(-1)
	}
	defer f.Close()

	scheme := ""
	addr := ""
	if m.config.Listen != "" {
		scheme = "http"
		addr = fullAddr(m.config.Listen)
	} else {
		scheme = "https"
		addr = fullAddr(m.config.TlsListen)
	}

	f.WriteString(fmt.Sprintf("api=%s://%s", scheme, addr))
}

func (m *apiServer) loadDConfig() error {
	confPath := rtclib.FullPath("conf/gortc.ini")

	config := &apiDConfig{}
	err := golib.ConfigFile(confPath, "APIModule", config)
	if err != nil {
		return fmt.Errorf("Parse dconfig %s Failed, %s", confPath, err)
	}
	m.dconfig = config

	return nil
}

func (m *apiServer) loadConfig() error {
	confPath := rtclib.FullPath("conf/gortc.ini")

	config := &apiConfig{}
	err := golib.ConfigFile(confPath, "APIModule", config)
	if err != nil {
		return fmt.Errorf("Parse config %s Failed, %s", confPath, err)
	}
	m.config = config

	return nil
}

func (m *apiServer) initLog() error {
	logPath := rtclib.FullPath(m.dconfig.LogFile)
	m.logLevel = golib.LoglvEnum.ConfEnum(m.dconfig.LogLevel, golib.LOGINFO)
	m.log = golib.NewLog(logPath)

	return nil
}

// api handler

func (m *apiServer) parseUri(uri string) (bool, string, string, string) {
	reg := regexp.MustCompile(`^/(\w+)/(v\d+)/(.+)`)

	match := reg.FindStringSubmatch(uri)
	if len(match) == 0 {
		return false, "", "", ""
	}

	return true, match[1], match[2], match[3]
}

func (m *apiServer) callAPI(req *http.Request, apiname string, version string,
	paras string) (int, *map[string]string, interface{},
	*map[int]rtclib.RespCode) {

	// API Call Access Phase
	// access.v1 is a special api for access auth
	// if load in apiserver, apiserver will call it Post interface first
	// if auth successd, go on to call indicated api
	// otherwise denied the api request
	api := am.getAPI("access.v1")
	if api != nil {
		code, _, msg, _ := api.Post(req, paras)
		if code != 0 { // code == 0, auth successd, otherwise, auth failed
			m.LogError("access failed, code %d, msg: %v", code, msg)
			return 7, nil, nil, nil
		}
	}

	// API Call Content Phase
	api = am.getAPI(apiname + "." + version)
	if api == nil {
		return 3, nil, nil, nil
	}

	switch req.Method {
	case "GET":
		return api.Get(req, paras)
	case "POST":
		return api.Post(req, paras)
	case "DELETE":
		return api.Delete(req, paras)
	}

	return 2, nil, nil, nil
}

func (m *apiServer) handler(w http.ResponseWriter, req *http.Request) {
	ok, apiname, version, paras := m.parseUri(req.URL.Path)
	if !ok {
		newResponse(1, nil, nil, nil).sendResp(w)
		return
	}

	newResponse(m.callAPI(req, apiname, version, paras)).sendResp(w)
}

// for module interface

func (m *apiServer) PreInit() error {
	if err := m.loadConfig(); err != nil {
		return err
	}

	if err := m.Reload(); err != nil {
		return err
	}

	golib.AddReloader("apiserver", m)

	return nil
}

func (m *apiServer) Init() error {
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

		s, err := golib.NewHTTPServer(m.config.TlsListen, m.config.Cert,
			m.config.Key, "/", m.dconfig.ClientHeaderTimeout,
			m.dconfig.Keepalived, m.log, m.handler, accessFile)
		if err != nil {
			return fmt.Errorf("New API TLSServer error: %s", err)
		}

		m.tlsServer = s

		m.nServers++
	}

	m.writeAPIFile()

	return nil
}

func (m *apiServer) PreMainloop() error {
	return nil
}

func (m *apiServer) Mainloop() {
	quit := make(chan bool)

	if m.server != nil {
		m.LogInfo("api server start ...")
		go func() {
			retry := 0
			for {
				err := m.server.Start()
				if err != nil {
					retry++
					m.LogError("api server start failure %d, %s", retry, err)

					if retry >= 10 {
						m.LogError("api server start failure, system exit")
						os.Exit(1)
					}

					time.Sleep(500 * time.Millisecond)
				} else {
					m.LogError("api server close")
					quit <- true
					break
				}
			}
		}()
	}

	if m.tlsServer != nil {
		m.LogInfo("api tlsserver start ...")
		go func() {
			retry := 0
			for {
				err := m.tlsServer.Start()
				if err != nil {
					retry++
					m.LogError("api tlsserver start failure %d, %s", retry, err)

					if retry >= 10 {
						m.LogError("api tlsserver start failure, system exit")
						os.Exit(1)
					}

					time.Sleep(500 * time.Millisecond)
				} else {
					m.LogError("api tlsserver close")
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

func (m *apiServer) Reload() error {
	if err := m.loadDConfig(); err != nil {
		return err
	}

	if err := m.initLog(); err != nil {
		return err
	}

	return nil
}

func (m *apiServer) Exit() {
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

func (m *apiServer) Prefix() string {
	return "[api] " + strconv.Itoa(os.Getpid())
}

func (m *apiServer) Suffix() string {
	return ""
}

func (m *apiServer) LogLevel() int {
	return m.logLevel
}

// for log ctx

func (m *apiServer) LogDebug(format string, v ...interface{}) {
	m.log.LogDebug(m, format, v...)
}

func (m *apiServer) LogInfo(format string, v ...interface{}) {
	m.log.LogInfo(m, format, v...)
}

func (m *apiServer) LogError(format string, v ...interface{}) {
	m.log.LogError(m, format, v...)
}

func (m *apiServer) LogFatal(format string, v ...interface{}) {
	m.log.LogFatal(m, format, v...)
}
