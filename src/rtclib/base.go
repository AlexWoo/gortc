// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Type

package rtclib

type Module interface {
	LoadConfig() bool
	Init() bool
	Run()
	Exit()
}
