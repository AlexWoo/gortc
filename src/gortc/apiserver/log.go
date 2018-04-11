// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Log

package apiserver

import (
	"fmt"
	"os"
	"rtclib"
	"time"
)

var loglevel = []string{
	"debug",
	"info ",
	"error",
	"fatal"}

var loglvEnum = map[string]int{
	"debug": rtclib.LOGDEBUG,
	"info":  rtclib.LOGINFO,
	"error": rtclib.LOGERROR,
	"fatal": rtclib.LOGFATAL}

type RTCLogHandle struct {
}

var apilog *rtclib.Log

func (handle RTCLogHandle) LogPrefix(loglv int) string {
	timestr := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s %s [apiserver] %d",
		timestr, loglevel[loglv], os.Getpid())
}

func (handle RTCLogHandle) LogSuffix(loglv int) string {
	return ""
}

func initLog(config *APIServerConfig) {
	logPath := rtclib.RTCPATH + "/logs/api.log"
	logLevel := rtclib.ConfEnum(rtclib.LoglvEnum, config.LogLevel,
		rtclib.LOGINFO)

	rtclogHandle := RTCLogHandle{}
	apilog = rtclib.NewLog(rtclogHandle, logPath, logLevel,
		int64(config.LogRotateSize))
	if apilog == nil {
		os.Exit(1)
	}

	return
}

func LogDebug(format string, v ...interface{}) {
	apilog.LogDebug(format, v...)
}

func LogInfo(format string, v ...interface{}) {
	apilog.LogInfo(format, v...)
}

func LogError(format string, v ...interface{}) {
	apilog.LogError(format, v...)
}

func LogFatal(format string, v ...interface{}) {
	apilog.LogFatal(format, v...)
}
