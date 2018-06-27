// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	"fmt"
	"sync"

	uuid "github.com/satori/go.uuid"
)

type SLP interface {
	// Task start in normal stage, msg process interface
	Process(jsip *JSIP)

	// Task start in SLP loaded in gortc, msg process interface
	OnLoad(jsip *JSIP)

	// Create SLP ctx
	NewSLPCtx() interface{}
}

type Task struct {
	Name string
	SLP  SLP
	ctx  interface{}
	dlgs []string
	lock sync.Mutex
	msgs chan *JSIP
	quit bool

	Process func(jsip *JSIP)
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

func (t *Task) GetCtx() interface{} {
	return t.ctx
}

func (t *Task) SetCtx(ctx interface{}) {
	t.ctx = ctx
}

func (t *Task) DelTask() {
	t.lock.Lock()
	defer t.lock.Unlock()

	taskRWLock.Lock()
	defer taskRWLock.Unlock()

	for _, dlg := range t.dlgs {
		delete(tasks, dlg)
	}
}

func (t *Task) run() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("task process err", err)
		}
		t.DelTask()
	}()

	for {
		select {
		case msg := <-t.msgs:
			t.Process(msg)
		}

		if t.quit {
			return
		}
	}
}

func (t *Task) SetFinished() {
	t.quit = true
}

func (t *Task) OnMsg(jsip *JSIP) {
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

	if dlg != "" {
		t.lock.Lock()
		t.dlgs = append(t.dlgs, dlg)
		t.lock.Unlock()

		taskRWLock.Lock()
		tasks[dlg] = t
		taskRWLock.Unlock()
	}

	go t.run()

	return t
}
