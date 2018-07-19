// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package rtcmodule

import (
	"os"
	"strconv"

	"github.com/alexwoo/golib"
)

type RTCLogHandle struct {
}

var log *golib.Log

func (handle RTCLogHandle) Prefix() string {
	return "[rtc] " + strconv.Itoa(os.Getpid())
}

func (handle RTCLogHandle) Suffix() string {
	return ""
}

func initLog(config *RTCModuleConfig) {
	logPath := module.rtcpath + "/logs/rtc.log"
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	rtclogHandle := RTCLogHandle{}
	log = golib.NewLog(rtclogHandle, logPath, logLevel)
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
