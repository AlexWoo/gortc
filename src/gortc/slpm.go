// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Service Logical Processor manager

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"rtclib"
	"time"

	"github.com/tidwall/gjson"
)

const (
	SLPONLOAD = iota
	SLPPROCESS
)

type slpPlugin struct {
	name     string
	used     uint64
	using    uint64
	file     string
	time     time.Time
	ctx      interface{}
	instance func(task *rtclib.Task) rtclib.SLP
}

type slpm struct {
	slps    map[string]*slpPlugin
	slpconf string
	slpdir  string
	plugins map[string]string
}

var sm *slpm

func slpmInstance() *slpm {
	if sm != nil {
		return sm
	}

	sm = &slpm{
		slps:    make(map[string]*slpPlugin),
		plugins: make(map[string]string),
	}

	return sm
}

// for module interface

func (m *slpm) PreInit() error {
	return nil
}

func (m *slpm) Init() error {
	m.slpconf = rtclib.FullPath("conf/.slps")
	m.slpdir = rtclib.FullPath("slp/")

	f, err := os.Open(m.slpconf)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", m.slpconf, err)
	}

	json, _ := ioutil.ReadAll(f)
	if !gjson.ValidBytes(json) {
		return fmt.Errorf("parse file %s failed: %v", m.slpconf, err)
	}

	j, ok := gjson.ParseBytes(json).Value().(map[string]interface{})
	if !ok {
		return fmt.Errorf("slp file %s format error: %v", m.slpconf, err)
	}

	for name, v := range j {
		path, ok := v.(string)
		if !ok {
			return fmt.Errorf("slp %s format error: %v", name, err)
		}

		if err := m.slpLoad(name, path); err != nil {
			return err
		}
		m.plugins[name] = path
	}

	return nil

}

func (m *slpm) PreMainloop() error {
	am.addInternalAPI("slpm.v1", Slpmv1)

	return nil
}

func (m *slpm) Mainloop() {
}

func (m *slpm) Reload() error {
	return nil
}

func (m *slpm) Reopen() error {
	return nil
}

func (m *slpm) Exit() {
}

// internal interface

func (m *slpm) updateSLPFile() error {
	f, err := os.OpenFile(m.slpconf, os.O_TRUNC|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		return fmt.Errorf("open file %s failed: %v", m.slpconf, err)
	}

	d, _ := json.Marshal(m.plugins)

	_, err = f.Write(d)
	if err != nil {
		return fmt.Errorf("write file %s failed: %v", m.slpconf, err)
	}

	return nil
}

func (m *slpm) slpLoad(name string, slpFile string) error {
	slp := &slpPlugin{
		name: name,
		file: slpFile,
		time: time.Now(),
	}
	path := m.slpdir + slpFile

	p, err := plugin.Open(path)
	if err != nil {
		return fmt.Errorf("open slp plugin error: %v", err)
	}

	v, err := p.Lookup("GetInstance")
	if err != nil {
		return fmt.Errorf("find slp plugin entry error: %v", err)
	}

	defer func() {
		if err := recover(); err != nil {
			rtcs.LogError("load %s %s failed: %v", name, path, err)
			if m.plugins[name] == "" {
				delete(m.plugins, name)
				m.updateSLPFile()
			}
		}
	}()

	slp.instance = v.(func(task *rtclib.Task) rtclib.SLP)
	m.slps[name] = slp

	// SLP Init Process when loaded
	t := rtclib.NewTask("", rtcs.taskQ, rtcs.log, rtcs.logLevel)
	t.Name = name
	m.getSLP(t, SLPONLOAD)
	if t.SLP == nil {
		return fmt.Errorf("get slp %s failed", name)
	}

	t.OnMsg(nil)

	return nil
}

func (m *slpm) addSLP(name string, slpFile string) string {
	if err := m.slpLoad(name, slpFile); err != nil {
		return fmt.Sprintf("Load SLP %s %s failed\n", name, slpFile)
	}

	m.plugins[name] = slpFile

	err := m.updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			m.slpconf, err)
	}

	return fmt.Sprintf("Load SLP %s %s successd\n", name, slpFile)
}

func (m *slpm) delSLP(name string) string {
	delete(m.plugins, name)
	delete(m.slps, name)

	err := m.updateSLPFile()
	if err != nil {
		return fmt.Sprintf("Update slp conf file %s error: %v\n",
			m.slpconf, err)
	}

	return fmt.Sprintf("Delete SLP %s successd\n", name)
}

func (m *slpm) getSLP(t *rtclib.Task, stage int) {
	p := m.slps[t.Name]
	if p == nil {
		rtcs.LogError("SLP %s not exist", t.Name)
		return
	}
	p.using++

	t.SLP = p.instance(t)
	if t.SLP == nil {
		rtcs.LogError("get slp error")
		return
	}

	switch stage {
	case SLPONLOAD:
		t.Process = t.SLP.OnLoad
		p.ctx = t.SLP.NewSLPCtx()
	case SLPPROCESS:
		t.Process = t.SLP.Process
	}

	t.SetCtx(p.ctx)
}

func (m *slpm) endSLP(t *rtclib.Task) {
	p := m.slps[t.Name]
	if p == nil { // SLP has been deleted
		return
	}

	p.using--
	p.used++
}

func (m *slpm) listSLP() string {
	ret := "slp\t\tused\t\tusing\t\tfile\t\ttime\n"
	ret += "------------------------------------------------------------\n"
	for _, v := range m.slps {
		ret += fmt.Sprintf("%s\t%d\t%d\t%s\t%s\n", v.name, v.used, v.using,
			v.file, v.time.Format("2006-01-02 15:04:05.000"))
	}
	ret += "------------------------------------------------------------\n"

	return ret
}
