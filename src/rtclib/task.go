// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	"sync"

	uuid "github.com/satori/go.uuid"
)

const (
	CONTINUE = iota
	FINISH
)

type SlpCtx struct {
	Body interface{}
}

type SLP interface {
	Process(jsip *JSIP) int
}

type Task struct {
	Name string
	SLP  SLP
	Ctx  *SlpCtx
	dlgs []string
	lock sync.Mutex
}

var tasks = make(map[string]*Task)

func (t *Task) NewDialogueID() string {
	u1, _ := uuid.NewV4()
	dlg := jstack.config.Realm + u1.String()

	t.lock.Lock()
	t.dlgs = append(t.dlgs, dlg)
	t.lock.Unlock()
	tasks[dlg] = t

	return dlg
}

func (t *Task) DelTask() {
	t.lock.Lock()
	defer t.lock.Unlock()
	for _, dlg := range t.dlgs {
		JsessDel(dlg)
		delete(tasks, dlg)
	}
}

func (t *Task) Process(jsip *JSIP) {
	if t.SLP.Process(jsip) == FINISH {
		t.DelTask()
	}
}

func GetTask(dlg string) *Task {
	return tasks[dlg]
}

func NewTask(dlg string) *Task {
	t := &Task{}
	t.lock.Lock()
	t.dlgs = append(t.dlgs, dlg)
	t.lock.Unlock()
	tasks[dlg] = t

	return t
}
