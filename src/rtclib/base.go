// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

import "github.com/alexwoo/golib"

var RTCPATH string

type Module interface {
	LoadConfig() bool
	Init(log *golib.Log) bool
	Run()
	Exit()
}
