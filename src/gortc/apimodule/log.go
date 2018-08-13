// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// API Log

package apimodule

import (
	"os"
	"strconv"

	"github.com/alexwoo/golib"
)

type RTCLogHandle struct {
}

type logctx struct {
	logLevel int
	log      *golib.Log
}

func (ctx *logctx) Prefix() string {
	return "[api] " + strconv.Itoa(os.Getpid())
}

func (ctx *logctx) Suffix() string {
	return ""
}

func (ctx *logctx) LogLevel() int {
	return ctx.logLevel
}

var apilogCtx *logctx

func initLog(config *APIModuleConfig, log *golib.Log) {
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	var logFile string
	if config.LogFile != "" {
		logFile = module.rtcpath + config.LogFile
	}

	if logFile == "" || logFile == log.LogPath() {
		apilogCtx = &logctx{
			logLevel: logLevel,
			log:      log,
		}
	} else {
		apilogCtx = &logctx{
			logLevel: logLevel,
			log:      golib.NewLog(logFile),
		}
	}

	if apilogCtx.log == nil {
		os.Exit(1)
	}

	return
}

func LogDebug(format string, v ...interface{}) {
	apilogCtx.log.LogDebug(apilogCtx, format, v...)
}

func LogInfo(format string, v ...interface{}) {
	apilogCtx.log.LogInfo(apilogCtx, format, v...)
}

func LogError(format string, v ...interface{}) {
	apilogCtx.log.LogError(apilogCtx, format, v...)
}

func LogFatal(format string, v ...interface{}) {
	apilogCtx.log.LogFatal(apilogCtx, format, v...)
}
