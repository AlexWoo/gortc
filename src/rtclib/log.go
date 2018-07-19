// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Base Log

package rtclib

import (
	"fmt"
	"os"
	"time"

	"github.com/alexwoo/golib"
)

const (
	LOGDEBUG = iota
	LOGINFO
	LOGERROR
	LOGFATAL
)

var LogLevel = []string{
	"debug",
	"info ",
	"error",
	"fatal",
}

var LoglvEnum = golib.Enum{
	"debug": LOGDEBUG,
	"info":  LOGINFO,
	"error": LOGERROR,
	"fatal": LOGFATAL,
}

type LogHandle interface {
	LogPrefix(loglv int) string
	LogSuffix(loglv int) string
}

type Log struct {
	logPath    string
	logLevel   int
	rotateSize int64
	logFile    *os.File
	logSize    int64

	handle LogHandle
}

func (log *Log) logPrintf(loglv int, format string, v ...interface{}) {
	len, err := fmt.Fprintf(log.logFile, log.handle.LogPrefix(loglv)+" "+format+
		" "+log.handle.LogSuffix(loglv)+"\n", v...)
	if err != nil {
		return
	}

	log.logSize += int64(len)
	log.logRotate()
}

func (log *Log) logRotate() {
	if log.logSize < log.rotateSize {
		return
	}

	backup := log.logPath + time.Now().Format(".20060102150405")
	os.Rename(log.logPath, backup)
	log.logFile.Close()

	log.logFile, _ = os.OpenFile(log.logPath,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	log.logSize = 0
}

func NewLog(handle LogHandle, logPath string, logLevel int,
	rotateSize int64) *Log {

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil
	}

	len, _ := f.Seek(0, os.SEEK_END)

	log := &Log{
		logPath:    logPath,
		logLevel:   logLevel,
		rotateSize: rotateSize,
		logFile:    f,
		logSize:    len,
		handle:     handle}

	log.logRotate()

	return log
}

func (log *Log) LogDebug(format string, v ...interface{}) {
	if log.logLevel > LOGDEBUG {
		return
	}

	log.logPrintf(LOGDEBUG, format, v...)
}

func (log *Log) LogInfo(format string, v ...interface{}) {
	if log.logLevel > LOGINFO {
		return
	}

	log.logPrintf(LOGINFO, format, v...)
}

func (log *Log) LogError(format string, v ...interface{}) {
	if log.logLevel > LOGERROR {
		return
	}

	log.logPrintf(LOGERROR, format, v...)
}

// Process will exit when call LogFatal
func (log *Log) LogFatal(format string, v ...interface{}) {
	if log.logLevel > LOGFATAL {
		return
	}

	log.logPrintf(LOGFATAL, format, v...)

	os.Exit(1)
}
