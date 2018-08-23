// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Log

package rtcmodule

import (
	"os"
	"rtclib"
	"strconv"

	"github.com/alexwoo/golib"
)

type logctx struct {
	logLevel int
	log      *golib.Log
}

func (ctx *logctx) Prefix() string {
	return "[rtc] " + strconv.Itoa(os.Getpid())
}

func (ctx *logctx) Suffix() string {
	return ""
}

func (ctx *logctx) LogLevel() int {
	return ctx.logLevel
}

var rtclogCtx *logctx

func initLog(config *RTCModuleConfig, log *golib.Log) {
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	logFile := rtclib.FullPath(config.LogFile)

	if logFile == "" || logFile == log.LogPath() {
		rtclogCtx = &logctx{
			logLevel: logLevel,
			log:      log,
		}
	} else {
		rtclogCtx = &logctx{
			logLevel: logLevel,
			log:      golib.NewLog(logFile),
		}
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
