// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// distribute Module

package main

import (
	"fmt"
	"rtclib"
	"sync"
)

type distribute struct {
	dlgTasksRWLock sync.RWMutex
	dlgtasks       map[string]*rtclib.Task
	relTasksRWLock sync.RWMutex
	reltasks       map[string]*rtclib.Task
	taskQ          chan *rtclib.Task
	exit           chan bool
}

var dist *distribute

func distInstance() *distribute {
	if dist != nil {
		return dist
	}

	dist = &distribute{
		dlgtasks: make(map[string]*rtclib.Task),
		reltasks: make(map[string]*rtclib.Task),
		taskQ:    make(chan *rtclib.Task),
		exit:     make(chan bool),
	}

	return dist
}

func (m *distribute) State() string {
	// dlgtasks
	ret := "dlgtasks:{\n"
	m.dlgTasksRWLock.RLock()
	for k, v := range m.dlgtasks {
		ret += fmt.Sprintf("    %s: %p\n", k, v)
	}
	m.dlgTasksRWLock.RUnlock()
	ret += "}\n"

	// reltasks
	ret += "reltasks:{\n"
	m.relTasksRWLock.RLock()
	for k, v := range m.reltasks {
		ret += fmt.Sprintf("    %s: %p\n", k, v)
	}
	m.relTasksRWLock.RUnlock()
	ret += "}\n"

	return ret
}

func (m *distribute) setdlg(dlg string, task *rtclib.Task) {
	if dlg == "" || task == nil {
		return
	}

	m.dlgTasksRWLock.Lock()
	defer m.dlgTasksRWLock.Unlock()

	m.dlgtasks[dlg] = task
}

func (m *distribute) process(jsip *rtclib.JSIP) {
	dlg := jsip.DialogueID

	m.dlgTasksRWLock.RLock()
	task := m.dlgtasks[dlg]
	if task != nil {
		task.OnMsg(jsip)
		m.dlgTasksRWLock.RUnlock()
		return
	}
	m.dlgTasksRWLock.RUnlock()

	if jsip.Code > 0 {
		rtcs.LogError("Receive %s but SLP is finished", jsip.Name())
		return
	}

	if jsip.Type != rtclib.INVITE && jsip.Type != rtclib.REGISTER &&
		jsip.Type != rtclib.OPTIONS && jsip.Type != rtclib.MESSAGE &&
		jsip.Type != rtclib.SUBSCRIBE {

		rtcs.LogError("Receive %s but SLP is finished", jsip.Name())
		return
	}

	slpname := "default"
	relid := ""

	if len(jsip.Router) != 0 {
		jsipUri, err := rtclib.ParseJSIPUri(jsip.Router[0])
		if err != nil {
			rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 400))
			return
		}

		rid, ok := jsipUri.Paras["relid"].(string)
		if ok && rid != "" {
			relid = rid
			m.relTasksRWLock.RLock()
			task := m.reltasks[relid]
			m.relTasksRWLock.RUnlock()
			if task != nil {
				m.setdlg(dlg, task)
				task.OnMsg(jsip)
				return
			}
		}

		name, ok := jsipUri.Paras["type"].(string)
		if ok && name != "" {
			slpname = name
		}
	}

	task = rtclib.NewTask(relid, m.taskQ, m.setdlg, rtcs.log, rtcs.logLevel)
	task.Name = slpname
	sm.getSLP(task, SLPPROCESS)
	if task.SLP == nil {
		rtclib.SendMsg(rtclib.JSIPMsgRes(jsip, 404))
		return
	}

	m.setdlg(dlg, task)
	if relid != "" {
		m.relTasksRWLock.Lock()
		m.reltasks[relid] = task
		m.relTasksRWLock.Unlock()
	}
	task.OnMsg(jsip)
}

func (m *distribute) delTask(task *rtclib.Task) {
	dlgs := task.GetDlgs()
	m.dlgTasksRWLock.Lock()
	for _, dlg := range dlgs {
		delete(m.dlgtasks, dlg)
	}
	m.dlgTasksRWLock.Unlock()

	relid := task.GetRelid()
	if relid != "" {
		delete(m.reltasks, relid)
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

func (m *distribute) Reload() error {
	return nil
}

func (m *distribute) Reopen() error {
	return nil
}

func (m *distribute) Exit() {
	m.exit <- true
}
