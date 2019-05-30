// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// distribute Module

package main

import (
	"fmt"
	"rtclib"
	"strings"
	"sync"
)

type distribute struct {
	relLock sync.RWMutex
	relids  map[string]*rtclib.Task
	taskQ   chan *rtclib.Task
	exit    chan bool
}

var dist *distribute

func distInstance() *distribute {
	if dist != nil {
		return dist
	}

	dist = &distribute{
		relids: make(map[string]*rtclib.Task),
		taskQ:  make(chan *rtclib.Task),
		exit:   make(chan bool),
	}

	return dist
}

func (m *distribute) State() string {
	m.relLock.RLock()
	defer m.relLock.RUnlock()

	ret := "relids:{\n"
	for k, v := range m.relids {
		ret += fmt.Sprintf("    %s: %p\n", k, v)
	}
	ret += "}\n"

	return ret
}

func (m *distribute) setRelated(id string, task *rtclib.Task) {
	if id == "" || task == nil {
		return
	}

	m.relLock.Lock()
	defer m.relLock.Unlock()

	m.relids[id] = task
}

func (m *distribute) getSrvNameByUri(uri string) string {
	jsipUri, err := rtclib.ParseJSIPUri(uri)
	if err != nil {
		return "default"
	}

	service := strings.Split(jsipUri.Host, ".")[0]

	p := sm.getSLPByName(service)
	if p == nil {
		return "default"
	}

	return service
}

func (m *distribute) process(jsip *rtclib.JSIP) {
	dlg := jsip.DialogueID

	// old DialogueID
	m.relLock.RLock()
	task := m.relids[dlg]
	if task != nil {
		if jsip.Type == rtclib.TERM {
			delete(m.relids, dlg)
		}
		m.relLock.RUnlock()

		task.OnMsg(jsip)

		return
	}
	m.relLock.RUnlock()

	// new request

	if jsip.Code > 0 {
		rtcs.LogError("Receive %s but SLP is finished", jsip.Name())
		return
	}

	if jsip.Type == rtclib.TERM {
		return
	}

	if jsip.Type != rtclib.INVITE && jsip.Type != rtclib.REGISTER &&
		jsip.Type != rtclib.OPTIONS && jsip.Type != rtclib.MESSAGE &&
		jsip.Type != rtclib.SUBSCRIBE {

		rtcs.LogError("Receive %s but SLP is finished", jsip.Name())
		return
	}

	slpname := "default"

	// get relid
	if len(jsip.Router) > 0 {
		jsipUri, err := rtclib.ParseJSIPUri(jsip.Router[0])
		if err != nil {
			rtcs.LogError("Router[%s] format error, %v", jsip.Router[0], err)
			rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 400))
			return
		}

		rid, ok := jsipUri.Paras["relid"].(string)
		if ok && rid != "" {
			// If has relid, but cannot find task, task has finished
			m.relLock.RLock()
			task = m.relids[rid]
			m.relLock.RUnlock()

			if task == nil {
				rtcs.LogError("Cannot find task for slp %s", slpname)
				rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 404))
				return
			}

			m.setRelated(dlg, task)

			task.OnMsg(jsip)
			return
		}

		name, ok := jsipUri.Paras["type"].(string)
		if ok && name != "" {
			slpname = name
		} else {
			slpname = m.getSrvNameByUri(jsipUri.Host)
		}
	} else {
		slpname = m.getSrvNameByUri(jsip.RequestURI)
	}

	// get task by slpname
	task = rtclib.NewTask(m.taskQ, m.setRelated, rtcs.log, rtcs.logLevel)
	task.Name = slpname
	sm.getSLP(task, SLPPROCESS)
	if task.SLP == nil {
		rtcs.LogError("Cannot find task for slp %s", slpname)
		rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 404))
		return
	}

	m.setRelated(dlg, task)

	task.OnMsg(jsip)
}

func (m *distribute) delTask(task *rtclib.Task) {
	ids := task.GetRelids()

	m.relLock.Lock()
	defer m.relLock.Unlock()

	for _, id := range ids {
		delete(m.relids, id)
	}
}

func (m *distribute) PreInit() error {
	return nil
}

func (m *distribute) Init() error {
	return nil
}

func (m *distribute) PreMainloop() error {
	return nil
}

func (m *distribute) Mainloop() {
	jsipC := rtclib.JStackInstance().JSIPChannel()

	for {
		select {
		case jsip := <-jsipC:
			m.process(jsip)
		case task := <-m.taskQ:
			m.delTask(task)
		case <-m.exit:
			return
		}
	}
}

func (m *distribute) Exit() {
	m.exit <- true
}
