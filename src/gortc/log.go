// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package main

import (
	"fmt"
	"os"
	"rtclib"
	"time"
)

type RTCLogHandle struct {
}

var rtclog *rtclib.Log

func (handle RTCLogHandle) LogPrefix(loglv int) string {
	timestr := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s %s [main] %d", timestr, rtclib.LogLevel[loglv],
		os.Getpid())
}

func (handle RTCLogHandle) LogSuffix(loglv int) string {
	return ""
}

func initLog() {
	logPath := rtcpath + "/logs/error.log"
	logLevel := rtclib.ConfEnum(rtclib.LoglvEnum, config.LogLevel,
		rtclib.LOGINFO)

	rtclogHandle := RTCLogHandle{}
	rtclog = rtclib.NewLog(rtclogHandle, logPath, logLevel,
		int64(config.LogRotateSize))
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
