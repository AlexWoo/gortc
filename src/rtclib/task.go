// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	"fmt"
	"sync"

	"github.com/alexwoo/golib"
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

	log      *golib.Log
	logLevel int

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

func NewTask(dlg string, taskq chan *Task, log *golib.Log, logLevel int) *Task {
	t := &Task{
		msgs:     make(chan *JSIP, 1024),
		taskq:    taskq,
		quitC:    make(chan bool, 1),
		log:      log,
		logLevel: logLevel,
	}

	t.SetDlg(dlg)

	go t.run()

	return t
}

func (t *Task) Prefix() string {
	return "[SLP]"
}

func (t *Task) Suffix() string {
	suf := ", " + t.Name
	if t.SLP != nil {
		suf += fmt.Sprintf(" %p", t.SLP)
	}
	return suf
}

func (t *Task) LogLevel() int {
	return t.logLevel
}

func (t *Task) LogDebug(format string, v ...interface{}) {
	t.log.LogDebug(t, format, v...)
}

func (t *Task) LogInfo(format string, v ...interface{}) {
	t.log.LogInfo(t, format, v...)
}

func (t *Task) LogError(format string, v ...interface{}) {
	t.log.LogError(t, format, v...)
}

func (t *Task) LogFatal(format string, v ...interface{}) {
	t.log.LogFatal(t, format, v...)
}
