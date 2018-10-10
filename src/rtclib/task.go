// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	"fmt"

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
	Name   string
	SLP    SLP
	ctx    interface{}
	dlgs   map[string]func(jsip *JSIP)
	relid  string
	msgs   chan *JSIP
	taskq  chan *Task
	quit   bool
	setdlg func(dlg string, task *Task)

	log      *golib.Log
	logLevel int

	Process func(jsip *JSIP)
}

func NewTask(relid string, taskq chan *Task,
	setdlg func(dlg string, task *Task), log *golib.Log, logLevel int) *Task {

	t := &Task{
		dlgs:     make(map[string]func(jsip *JSIP)),
		relid:    relid,
		msgs:     make(chan *JSIP, 1024),
		taskq:    taskq,
		setdlg:   setdlg,
		log:      log,
		logLevel: logLevel,
	}

	go t.run()

	return t
}

// New a jsip DialogueID for sending a new jsip session
func (t *Task) NewDialogueID() string {
	u1, _ := uuid.NewV4()
	dlg := jstack.dconfig.Realm + u1.String()

	t.dlgs[dlg] = nil
	t.setdlg(dlg, t)

	return dlg
}

// New a jsip DialogueID with jsip msg process entry,
// for sending a new jsip session
func (t *Task) NewDialogueIDWithEntry(process func(*JSIP)) string {
	u1, _ := uuid.NewV4()
	dlg := jstack.dconfig.Realm + u1.String()

	t.dlgs[dlg] = process
	t.setdlg(dlg, t)

	return dlg
}

// Get SLP ctx
func (t *Task) GetCtx() interface{} {
	return t.ctx
}

// Terminate SLP instance
func (t *Task) SetFinished() {
	t.quit = true
}

func (t *Task) run() {
	defer func() {
		if err := recover(); err != nil {
			t.LogError("task process err: %s", err)
			t.taskq <- t
		}
	}()

	for {
		msg := <-t.msgs
		if msg == nil {
			t.Process(msg)
			continue
		}

		entry := t.dlgs[msg.DialogueID]
		if entry == nil {
			t.Process(msg)
		} else {
			entry(msg)
		}

		if t.quit {
			t.taskq <- t
			return
		}
	}
}

func (t *Task) SetCtx(ctx interface{}) {
	t.ctx = ctx
}

func (t *Task) OnMsg(jsip *JSIP) {
	t.msgs <- jsip
}

func (t *Task) GetDlgs() []string {
	dlgs := make([]string, len(t.dlgs))
	for dlg, _ := range t.dlgs {
		dlgs = append(dlgs, dlg)
	}

	return dlgs
}

func (t *Task) GetRelid() string {
	return t.relid
}

// for log ctx

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

// for log

// log a debug level log
func (t *Task) LogDebug(format string, v ...interface{}) {
	t.log.LogDebug(t, format, v...)
}

// log a info level log
func (t *Task) LogInfo(format string, v ...interface{}) {
	t.log.LogInfo(t, format, v...)
}

// log a error level log
func (t *Task) LogError(format string, v ...interface{}) {
	t.log.LogError(t, format, v...)
}

// log a fatal level log, it will cause system exit
func (t *Task) LogFatal(format string, v ...interface{}) {
	t.log.LogFatal(t, format, v...)
}
