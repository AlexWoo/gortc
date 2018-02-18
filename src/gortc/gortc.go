// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// MainModule

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/go-ini/ini"
)

func GetGID() uint64 {
	b := make([]byte, 64)
	b = b[:runtime.Stack(b, false)]
	b = bytes.TrimPrefix(b, []byte("goroutine "))
	b = b[:bytes.IndexByte(b, ' ')]
	n, _ := strconv.ParseUint(string(b), 10, 64)
	return n
}

type MainConfig struct {
	A bool   `default:"false"`
	B string `default:"Hello"`
	C uint64 `default:"99999"`
	D int64  `default:"-1000"`
	E Size_t `default:"100M"`
	F Msec_t `default:"10s"`
}

type MainModule struct {
	name    string
	signals chan os.Signal
	config  *MainConfig
}

func NewMainModule() *MainModule {
	m := new(MainModule)
	m.name = "main"
	m.signals = make(chan os.Signal)

	return m
}

func (m *MainModule) LoadConfig() bool {
	m.config = new(MainConfig)

	f, err := ini.Load(CONFPATH)
	if err != nil {
		//TODO logErr
		return false
	}

	return Config(f, "", m.config)
}

func (m *MainModule) Init() bool {
	// quit
	signal.Notify(m.signals, syscall.SIGQUIT)

	// terminate
	signal.Notify(m.signals, syscall.SIGTERM)

	// reconfigure
	signal.Notify(m.signals, syscall.SIGHUP)

	// ignore
	//m.signal.Ignore(syscall.SIGINT)
	signal.Ignore(syscall.SIGALRM)
	signal.Ignore(syscall.SIGUSR1)
	signal.Ignore(syscall.SIGUSR2)
	signal.Ignore(syscall.SIGPIPE)

	return true
}

func (m *MainModule) Run() {
	for {
		s := <-m.signals
		fmt.Println("get signal: ", s)
	}
}

func (m *MainModule) State() {
}

func (m *MainModule) Exit() {
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	AddModule("API server", NewAPIServerModule())
}

func main() {
	AddModule("main", NewMainModule())

	ConfigMdoule()
	InitModule()

	RunModule()
}
