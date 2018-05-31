// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	"sync"

	uuid "github.com/satori/go.uuid"
)

type SlpCtx struct {
	Body interface{}
}

type SLP interface {
	Process(jsip *JSIP)
}

type Task struct {
	Name string
	SLP  SLP
	Ctx  *SlpCtx
	dlgs []string
	lock sync.Mutex
	msgs chan *JSIP
	quit bool
}

var tasks = make(map[string]*Task)
var taskRWLock sync.RWMutex

func (t *Task) NewDialogueID() string {
	u1, _ := uuid.NewV4()
	dlg := jstack.config.Realm + u1.String()

	t.lock.Lock()
	t.dlgs = append(t.dlgs, dlg)
	t.lock.Unlock()

	taskRWLock.Lock()
	tasks[dlg] = t
	taskRWLock.Unlock()

	return dlg
}

func (t *Task) DelTask() {
	t.lock.Lock()
	defer t.lock.Unlock()

	taskRWLock.Lock()
	defer taskRWLock.Unlock()

	for _, dlg := range t.dlgs {
		JsessDel(dlg)
		delete(tasks, dlg)
	}
}

func (t *Task) run() {
	defer t.DelTask()

	for {
		select {
		case msg := <-t.msgs:
			t.SLP.Process(msg)
		}

		if t.quit {
			return
		}
	}
}

func (t *Task) SetFinished() {
	t.quit = true
}

func (t *Task) Process(jsip *JSIP) {
	t.msgs <- jsip
}

func GetTask(dlg string) *Task {
	taskRWLock.RLock()
	defer taskRWLock.RUnlock()

	return tasks[dlg]
}

func NewTask(dlg string) *Task {
	t := &Task{
		msgs: make(chan *JSIP, 32),
	}

	t.lock.Lock()
	t.dlgs = append(t.dlgs, dlg)
	t.lock.Unlock()

	taskRWLock.Lock()
	tasks[dlg] = t
	taskRWLock.Unlock()

	go t.run()

	return t
}
