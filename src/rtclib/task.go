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
	Name  string
	SLP   SLP
	ctx   interface{}
	dlgs  []string
	msgs  chan *JSIP
	taskq chan *Task
	quitC chan bool
	quit  bool

	Process func(jsip *JSIP)
}

var tasks = make(map[string]*Task)
var taskRWLock sync.RWMutex

func (t *Task) NewDialogueID() string {
	u1, _ := uuid.NewV4()
	dlg := jstack.config.Realm + u1.String()

	t.SetDlg(dlg)

	return dlg
}

func (t *Task) GetCtx() interface{} {
	return t.ctx
}

func (t *Task) SetCtx(ctx interface{}) {
	t.ctx = ctx
}

func (t *Task) DelTask() {
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
			t.taskq <- t
			<-t.quitC
		}
	}()

	for {
		select {
		case msg := <-t.msgs:
			t.Process(msg)
			if t.quit {
				t.taskq <- t
			}
		case <-t.quitC:
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

func (t *Task) SetDlg(dlg string) {
	if dlg == "" {
		return
	}

	taskRWLock.Lock()
	defer taskRWLock.Unlock()

	t.dlgs = append(t.dlgs, dlg)
	tasks[dlg] = t
}

func GetTask(dlg string) *Task {
	taskRWLock.RLock()
	defer taskRWLock.RUnlock()

	return tasks[dlg]
}

func NewTask(dlg string, taskq chan *Task) *Task {
	t := &Task{
		msgs:  make(chan *JSIP, 1024),
		taskq: taskq,
		quitC: make(chan bool, 1),
	}

	t.SetDlg(dlg)

	go t.run()

	return t
}
