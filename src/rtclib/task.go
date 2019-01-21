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
	msgs  chan *JSIP
	taskq chan *Task
	quit  chan bool

	// DialogueID, RelID will save in relids table
	relids     map[string]func(jsip *JSIP)
	relLock    sync.RWMutex
	setRelated func(id string, task *Task)

	log      *golib.Log
	logLevel int

	Process func(jsip *JSIP)
}

func NewTask(taskq chan *Task, setRelated func(dlg string, task *Task),
	log *golib.Log, logLevel int) *Task {

	t := &Task{
		msgs:       make(chan *JSIP, 1024),
		taskq:      taskq,
		quit:       make(chan bool, 1),
		relids:     make(map[string]func(jsip *JSIP)),
		setRelated: setRelated,
		log:        log,
		logLevel:   logLevel,
	}

	go t.run()

	return t
}

// New a jsip DialogueID for sending a new jsip session
func (t *Task) NewDialogueID() string {
	u4, _ := uuid.NewV4()
	dlg := "dlg_" + jstack.dconfig.Realm + "_" + u4.String()

	t.relLock.Lock()
	t.relids[dlg] = nil
	t.relLock.Unlock()

	t.setRelated(dlg, t)

	return dlg
}

// New a jsip DialogueID with jsip msg process entry,
// for sending a new jsip session
func (t *Task) NewDialogueIDWithEntry(process func(*JSIP)) string {
	u4, _ := uuid.NewV4()
	dlg := "dlg_" + jstack.dconfig.Realm + "_" + u4.String()

	t.relLock.Lock()
	t.relids[dlg] = process
	t.relLock.Unlock()

	t.setRelated(dlg, t)

	return dlg
}

// New a relid with jsip msg process entry
// relid is used to trigger service by rtcbroker:
// 		relid fill in the second Router of request sent to service
// 		service send new request back to rtcbroker with second Router
//		gortc can send the request to rtcbroker instance by this relid
func (t *Task) NewRelIDWithEntry(process func(*JSIP)) string {
	u4, _ := uuid.NewV4()
	relid := "rel" + jstack.dconfig.Realm + "+" + u4.String()

	t.relLock.Lock()
	t.relids[relid] = process
	t.relLock.Unlock()

	t.setRelated(relid, t)

	return relid
}

// Get SLP ctx
func (t *Task) GetCtx() interface{} {
	return t.ctx
}

// Terminate SLP instance
func (t *Task) SetFinished() {
	t.quit <- true
}

func (t *Task) run() {
	defer func() {
		if err := recover(); err != nil {
			t.LogError("task process err: %s", err)
			t.taskq <- t
		}
	}()

	for {
		select {
		case msg := <-t.msgs:

			// Onload when slp load into system
			if msg == nil {
				t.Process(msg)
				continue
			}

			t.relLock.RLock()
			entry, ok := t.relids[msg.DialogueID]
			if ok { // old DialogueID
				if entry == nil {
					t.Process(msg)
				} else {
					entry(msg)
				}

				continue
			}
			t.relLock.RUnlock()

			// new DialogueID

			// get relid
			relid := ""
			if len(msg.Router) > 0 {
				jsipUri, _ := ParseJSIPUri(msg.Router[0])

				rid, ok := jsipUri.Paras["relid"].(string)
				if ok && rid != "" {
					relid = rid
				}
			}

			if relid != "" {
				t.relLock.Lock()
				entry = t.relids[relid]
				// New DialogueID use same entry with it's relid
				t.relids[msg.DialogueID] = entry
				t.relLock.Unlock()
			} else {
				t.relLock.RLock()
				entry = t.relids[msg.DialogueID]
				t.relLock.RUnlock()
			}

			if entry == nil {
				t.Process(msg)
			} else {
				entry(msg)
			}

		case <-t.quit:
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

func (t *Task) GetRelids() []string {
	t.relLock.RLock()
	defer t.relLock.RUnlock()

	relids := make([]string, len(t.relids))
	for id := range t.relids {
		relids = append(relids, id)
	}

	return relids
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
