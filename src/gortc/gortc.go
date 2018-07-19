// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// MainModule

package main

import (
	"bytes"
	"fmt"
	"gortc/apimodule"
	"gortc/rtcmodule"
	"os"
	"os/signal"
	"rtclib"
	"runtime"
	"strconv"
	"syscall"
	"time"

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
	LogLevel      string
	LogRotateSize rtclib.Size_t
}

var config *MainConfig
var signals chan os.Signal

func loadConfig() {
	confPath := rtclib.RTCPATH + "/conf/gortc.ini"
	config = new(MainConfig)

	f, err := ini.Load(confPath)
	if err != nil {
		fmt.Println("Load "+confPath+" Failed: ", err)
		os.Exit(1)
	}

	if !rtclib.Config(f, "", config) {
		fmt.Println("Parse config " + confPath + " Failed")
		os.Exit(1)
	}
}

func initSignal() {
	signals = make(chan os.Signal)

	// quit
	signal.Notify(signals, syscall.SIGQUIT)
	signal.Notify(signals, syscall.SIGINT)

	// terminate
	signal.Notify(signals, syscall.SIGTERM)

	// ignore
	signal.Ignore(syscall.SIGHUP)
	signal.Ignore(syscall.SIGALRM)
	signal.Ignore(syscall.SIGUSR1)
	signal.Ignore(syscall.SIGUSR2)
	signal.Ignore(syscall.SIGPIPE)
}

func mainloop() {
	runModule()
	t := time.NewTimer(time.Second * 1)
	exit := false

	for {
		select {
		case s := <-signals:
			fmt.Println("get signal: ", s)

			switch s {
			case syscall.SIGINT:
				exitModule()
			case syscall.SIGQUIT:
				exitModule()
			case syscall.SIGTERM:
				exit = true
			}
		case <-t.C:
			if !checkModule() {
				LogError("All Module Exit")
				exit = true
			}
			t = time.NewTimer(time.Second * 1)
		}

		if exit {
			break
		}
	}
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	addModule("apimodule", apimodule.NewAPIModule())
	addModule("rtcmodule", rtcmodule.NewRTCModule())
}

func main() {
	fmt.Println("gortc initSignal ...")
	initSignal()
	fmt.Println("gortc loadConfig ...")
	loadConfig()
	fmt.Println("gortc initLog ...")
	initLog()

	LogInfo("gortc init modules ...")
	initModule()
	apimodule.AddInternalAPI("runtime.v1", RunTimeV1)

	LogInfo("gortc init successd, start Running ...")

	mainloop()

	LogInfo("gortc gracefully stop")
}
