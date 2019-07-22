// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// ChatRoom demo
// JSIP Msg distribute

package main

import (
	"rtclib"
	"sync"
)

type ChatRoom struct {
	task *rtclib.Task
}

var (
	once    sync.Once
	manager *roomManager
)

func GetInstance(task *rtclib.Task) rtclib.SLP {
	slp := &ChatRoom{
		task: task,
	}

	// slp will receive TERM if dlg has terminate
	slp.task.TermNotify = true

	return slp
}

func (slp *ChatRoom) NewSLPCtx() interface{} {
	once.Do(func() {
		manager = slp.newRoomManager()
	})

	return manager
}

func (slp *ChatRoom) Process(msg *rtclib.JSIP) {
	if msg.Type == rtclib.TERM {
		slp.task.SetFinished()
		return
	}

	m := slp.task.GetCtx().(*roomManager)
	if m == nil {
		slp.task.LogError("Recv %s but roomManager not init", msg.Name())

		rtclib.SendMsg(rtclib.JSIPMsgRes(msg, 500))

		return
	}

	m.process(msg)
}

func (slp *ChatRoom) OnLoad(msg *rtclib.JSIP) {
}
