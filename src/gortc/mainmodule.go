// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// main Module

package main

import (
	"fmt"
	"os"
	"rtclib"
	"strconv"

	"github.com/alexwoo/golib"
)

type mainDConfig struct {
	LogFile  string `default:"logs/rtc.log"`
	LogLevel string `default:"info"`
}

type mainModule struct {
	dconfig  *mainDConfig
	log      *golib.Log
	logLevel int
}

func (m *mainModule) loadDConfig() error {
	confPath := rtclib.FullPath("conf/gortc.ini")

	config := &mainDConfig{}
	err := golib.ConfigFile(confPath, "", config)
	if err != nil {
		return fmt.Errorf("Parse dconfig %s Failed, %s", confPath, err)
	}
	m.dconfig = config

	return nil
}

func (m *mainModule) initLog() error {
	logPath := rtclib.FullPath(m.dconfig.LogFile)

	log := golib.NewLog(logPath)
	if log == nil {
		return fmt.Errorf("new log %s faild", logPath)
	}
	m.log = log

	m.logLevel = golib.LoglvEnum.ConfEnum(m.dconfig.LogLevel, golib.LOGINFO)

	return nil
}

// for module interface

func (m *mainModule) PreInit() error {
	if err := m.loadDConfig(); err != nil {
		return err
	}

	if err := m.initLog(); err != nil {
		return err
	}

	ms := golib.NewModules()
	ms.SetLog(m.log.Log())

	return nil
}

func (m *mainModule) Init() error {
	return nil
}

func (m *mainModule) PreMainloop() error {
	am.addInternalAPI("runtime.v1", RunTimeV1)

	return nil
}

func (m *mainModule) Mainloop() {
}

func (m *mainModule) Reload() error {
	if err := m.loadDConfig(); err != nil {
		return err
	}

	if err := m.initLog(); err != nil {
		return err
	}

	ms := golib.NewModules()
	ms.SetLog(m.log.Log())

	return nil
}

func (m *mainModule) Reopen() error {
	if err := m.initLog(); err != nil {
		return err
	}

	ms := golib.NewModules()
	ms.SetLog(m.log.Log())

	return nil
}

func (m *mainModule) Exit() {
}

// for log ctx

func (m *mainModule) Prefix() string {
	return "[main] " + strconv.Itoa(os.Getpid())
}

func (m *mainModule) Suffix() string {
	return ""
}

func (m *mainModule) LogLevel() int {
	return m.logLevel
}

// for log

func (m *mainModule) LogDebug(format string, v ...interface{}) {
	m.log.LogDebug(m, format, v...)
}

func (m *mainModule) LogInfo(format string, v ...interface{}) {
	m.log.LogInfo(m, format, v...)
}

func (m *mainModule) LogError(format string, v ...interface{}) {
	m.log.LogError(m, format, v...)
}

func (m *mainModule) LogFatal(format string, v ...interface{}) {
	m.log.LogFatal(m, format, v...)
}
