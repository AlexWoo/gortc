// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package rtcmodule

import (
	"os"
	"strconv"

	"github.com/alexwoo/golib"
)

type logctx struct {
	log *golib.Log
}

func (ctx *logctx) Prefix() string {
	return "[rtc] " + strconv.Itoa(os.Getpid())
}

func (ctx *logctx) Suffix() string {
	return ""
}

var rtclogCtx *logctx

func initLog(config *RTCModuleConfig) {
	logPath := module.rtcpath + "/logs/rtc.log"
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	rtclogCtx = &logctx{
		log: golib.NewLog(logPath, logLevel),
	}

	if rtclogCtx.log == nil {
		os.Exit(1)
	}

	return
}

func LogDebug(format string, v ...interface{}) {
	rtclogCtx.log.LogDebug(rtclogCtx, format, v...)
}

func LogInfo(format string, v ...interface{}) {
	rtclogCtx.log.LogInfo(rtclogCtx, format, v...)
}

func LogError(format string, v ...interface{}) {
	rtclogCtx.log.LogError(rtclogCtx, format, v...)
}

func LogFatal(format string, v ...interface{}) {
	rtclogCtx.log.LogFatal(rtclogCtx, format, v...)
}
