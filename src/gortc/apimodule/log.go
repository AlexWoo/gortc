// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Log

package apimodule

import (
	"fmt"
	"os"
	"rtclib"
	"time"
)

type RTCLogHandle struct {
}

var log *rtclib.Log

func (handle *RTCLogHandle) LogPrefix(loglv int) string {
	timestr := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s %s [api] %d",
		timestr, rtclib.LogLevel[loglv], os.Getpid())
}

func (handle *RTCLogHandle) LogSuffix(loglv int) string {
	return ""
}

func initLog(config *APIModuleConfig) {
	logPath := rtclib.RTCPATH + "/logs/api.log"
	logLevel := rtclib.ConfEnum(rtclib.LoglvEnum, config.LogLevel,
		rtclib.LOGINFO)

	rtclogHandle := &RTCLogHandle{}
	log = rtclib.NewLog(rtclogHandle, logPath, logLevel,
		int64(config.LogRotateSize))
	if log == nil {
		os.Exit(1)
	}

	return
}

func LogDebug(format string, v ...interface{}) {
	log.LogDebug(format, v...)
}

func LogInfo(format string, v ...interface{}) {
	log.LogInfo(format, v...)
}

func LogError(format string, v ...interface{}) {
	log.LogError(format, v...)
}

func LogFatal(format string, v ...interface{}) {
	log.LogFatal(format, v...)
}
