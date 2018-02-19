// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Server Module

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-ini/ini"
)

const (
	LOGDEBUG = iota
	LOGINFO
	LOGERROR
	LOGFATAL
)

var loglevel = []string{
	"debug",
	"info ",
	"error",
	"fatal"}

var loglvEnum = map[string]int{
	"debug": LOGDEBUG,
	"info":  LOGINFO,
	"error": LOGERROR,
	"fatal": LOGFATAL}

type LogHandle interface {
	LogPrintf(log *Log, loglv int, format string, v ...interface{})
}

type LogConfig struct {
	LogPath       string
	LogLevel      string
	LogRotateSize Size_t
}

type MainLogHandle struct {
}

func (handle *MainLogHandle) LogPrintf(log *Log, loglv int, format string,
	v ...interface{}) {

	timestr := time.Now().Format("2006-01-02 15:04:05.000")

	prefix := fmt.Sprintf("%s %s [%s] %d ", timestr, loglevel[loglv],
		log.module, os.Getpid())

	// YYYY-MM-DD hh:mm:ss.ms level module pid userstring
	len, err := fmt.Fprintf(log.logFile, prefix+format+"\n", v...)
	if err != nil {
		return
	}

	log.len += int64(len)
	log.rotateFile()
}

type Log struct {
	name     string
	module   string
	logFile  *os.File
	loglevel int
	len      int64

	config *LogConfig
	handle LogHandle
}

var MainLog *Log

func NewLog() *Log {
	MainLog = &Log{name: "Log", module: "main", handle: &MainLogHandle{}}

	return MainLog
}

func (log *Log) rotateFile() {
	if log.len < int64(log.config.LogRotateSize) {
		return
	}

	backup := log.config.LogPath + time.Now().Format(".20060102150405")
	os.Rename(log.config.LogPath, backup)

	defer log.logFile.Close()

	f, err := os.OpenFile(log.config.LogPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.LogError("Rotate log file %s Failed, %V", log.config.LogPath, err)
	}

	log.logFile = f
	log.len = 0
}

func (log *Log) LoadConfig() bool {
	log.config = new(LogConfig)

	f, err := ini.Load(CONFPATH)
	if err != nil {
		LogError("Load config file %s error: %v", CONFPATH, err)
		return false
	}

	return Config(f, "MainLog", log.config)
}

func (log *Log) Init() bool {
	if !strings.HasPrefix(log.config.LogPath, "/") &&
		!strings.HasPrefix(log.config.LogPath, "./") {

		log.config.LogPath = RTCPATH + log.config.LogPath
	}

	f, err := os.OpenFile(log.config.LogPath,
		os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		LogError("Open file %s Failed: %v", log.config.LogPath, err)
		return false
	}

	log.logFile = f
	len, _ := log.logFile.Seek(0, os.SEEK_END)
	if len >= int64(log.config.LogRotateSize) {
		log.rotateFile()
	} else {
		log.len = len
	}

	log.loglevel = ConfEnum(loglvEnum, log.config.LogLevel, LOGINFO)

	return true
}

func (log *Log) Run() {
}

func (log *Log) State() {
}

func (log *Log) Exit() {
}

func (log *Log) LogDebug(format string, v ...interface{}) {
	if log.loglevel > LOGDEBUG {
		return
	}

	log.handle.LogPrintf(log, LOGDEBUG, format, v...)
}

func (log *Log) LogInfo(format string, v ...interface{}) {
	if log.loglevel > LOGINFO {
		return
	}

	log.handle.LogPrintf(log, LOGINFO, format, v...)
}

func (log *Log) LogError(format string, v ...interface{}) {
	if log.loglevel > LOGERROR {
		return
	}

	log.handle.LogPrintf(log, LOGERROR, format, v...)
}

// Process will exit when call LogFatal
func (log *Log) LogFatal(format string, v ...interface{}) {
	if log.loglevel > LOGFATAL {
		return
	}

	log.handle.LogPrintf(log, LOGFATAL, format, v...)

	os.Exit(1)
}

func logStd(loglv int, format string, v ...interface{}) bool {
	if MainLog == nil || MainLog.logFile == nil {
		timestr := time.Now().Format("2006-01-02 15:04:05.000")

		prefix := fmt.Sprintf("%s %s [main] %d ", timestr, loglevel[loglv],
			os.Getpid())

		// YYYY-MM-DD hh:mm:ss.ms loglevel [main] pid userstring
		fmt.Fprintf(os.Stdout, prefix+format+"\n", v...)

		return true
	}

	return false
}

func LogDebug(format string, v ...interface{}) {
	if logStd(LOGDEBUG, format, v...) {
		return
	}

	MainLog.LogDebug(format, v...)
}

func LogInfo(format string, v ...interface{}) {
	if logStd(LOGINFO, format, v...) {
		return
	}

	MainLog.LogInfo(format, v...)
}

func LogError(format string, v ...interface{}) {
	if logStd(LOGERROR, format, v...) {
		return
	}

	MainLog.LogError(format, v...)
}

func LogFatal(format string, v ...interface{}) {
	if logStd(LOGFATAL, format, v...) {
		os.Exit(1)
		return
	}

	MainLog.LogFatal(format, v...)
}
