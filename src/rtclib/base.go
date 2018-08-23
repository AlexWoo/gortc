// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

import (
	"strings"

	"github.com/alexwoo/golib"
)

var RTCPATH string

type Module interface {
	LoadConfig() bool
	Init(log *golib.Log) bool
	Run()
	Exit()
}

func FullPath(path string) string {
	if path == "" {
		return path
	}

	if strings.HasSuffix(path, "/") {
		return path
	}

	if strings.HasSuffix(RTCPATH, "/") {
		return RTCPATH + path
	} else {
		return RTCPATH + "/" + path
	}
}
