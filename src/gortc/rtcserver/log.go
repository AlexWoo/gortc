// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package rtcserver

import (
	"fmt"
	"gortc/rtclib"
	"os"
	"strings"
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

var rtclog *rtclib.Log

func (handle RTCLogHandle) LogPrefix(loglv int) string {
	timestr := time.Now().Format("2006-01-02 15:04:05.000")
	return fmt.Sprintf("%s %s [rtcserver] %d",
		timestr, loglevel[loglv], os.Getpid())
}

func (handle RTCLogHandle) LogSuffix(loglv int) string {
	return ""
}

func initLog(config *RTCServerConfig, rtcPath string) {
	if !strings.HasPrefix(config.LogPath, "/") &&
		!strings.HasPrefix(config.LogPath, "./") {

		config.LogPath = rtcPath + config.LogPath
	}

	logLevel := rtclib.ConfEnum(loglvEnum, config.LogLevel, rtclib.LOGINFO)

	rtclogHandle := RTCLogHandle{}
	rtclog = rtclib.NewLog(rtclogHandle, config.LogPath, logLevel,
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
