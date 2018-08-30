// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// main func entry

package main

import (
	"rtclib"
	"runtime"

	"github.com/alexwoo/golib"
)

var (
	rtcpath = "/usr/local/gortc/"
)

func init() {
	rtclib.RTCPATH = rtcpath
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	ms := golib.NewModules()

	ms.AddModule("main", &mainModule{})
	ms.AddModule("apiserver", apiServerInstance())
	ms.AddModule("apimanager", apimInstance())
	ms.AddModule("rtcserver", rtcServerInstance())
	ms.AddModule("slpmanager", slpmInstance())

	ms.Start()
}
