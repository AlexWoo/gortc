// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package main

import (
	"os"
	"strconv"

	"github.com/alexwoo/golib"
)

type RTCLogHandle struct {
}

var rtclog *golib.Log

func (handle RTCLogHandle) Prefix() string {
	return "[main] " + strconv.Itoa(os.Getpid())
}

func (handle RTCLogHandle) Suffix() string {
	return ""
}

func initLog() {
	logPath := rtcpath + "/logs/error.log"
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	rtclogHandle := RTCLogHandle{}
	rtclog = golib.NewLog(rtclogHandle, logPath, logLevel)
	if rtclog == nil {
		os.Exit(1)
	}

	return
}

func LogDebug(format string, v ...interface{}) {
	rtclog.LogDebug(format, v...)
}

func LogInfo(format string, v ...interface{}) {
	rtclog.LogInfo(format, v...)
}

func LogError(format string, v ...interface{}) {
	rtclog.LogError(format, v...)
}

func LogFatal(format string, v ...interface{}) {
	rtclog.LogFatal(format, v...)
}
