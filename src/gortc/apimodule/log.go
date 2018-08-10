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
	log *golib.Log
}

func (ctx *logctx) Prefix() string {
	return "[api] " + strconv.Itoa(os.Getpid())
}

func (ctx *logctx) Suffix() string {
	return ""
}

var apilogCtx *logctx

func initLog(config *APIModuleConfig) {
	logPath := module.rtcpath + "/logs/api.log"
	logLevel := golib.LoglvEnum.ConfEnum(config.LogLevel, golib.LOGINFO)

	apilogCtx = &logctx{
		log: golib.NewLog(logPath, logLevel),
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
