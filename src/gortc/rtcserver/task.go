// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtcserver

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

const (
	CONTINUE = iota
	FINISH
)

type Task struct {
	name string
	slp  SLP
	dlgs []string
}

var tasks = make(map[string]*Task)

func (t *Task) newDialogueID() string {
	u1, _ := uuid.NewV1()
	dlg := rtcServerModule.config.Realm + u1.String()

	t.dlgs = append(t.dlgs, dlg)
	tasks[dlg] = t

	return dlg
}

func (t *Task) delTask() {
	for _, dlg := range t.dlgs {
		delete(tasks, dlg)
	}
}

func (t *Task) process(jsip *JSIP) {
	if t.slp.Process(jsip) == FINISH {
		t.delTask()
	}
}

func newTask(dlg string) *Task {
	t := &Task{}
	t.dlgs = append(t.dlgs, dlg)
	tasks[dlg] = t

	return t
}

func TaskProcess(jsip *JSIP) {
	dlg := jsip.DialogueID
	t := tasks[dlg]
	if t != nil {
		t.process(jsip)
		return
	}

	slpname := "default"

	if len(jsip.Router) != 0 {
		router0 := jsip.Router[0]
		_, _, paras := JsipParseUri(router0)

		for _, para := range paras {
			if strings.HasPrefix(para, "type=") {
				ss := strings.SplitN(para, "=", 2)
				if ss[1] != "" {
					slpname = ss[1]
				}
			}
		}
	}

	t = newTask(dlg)
	t.name = slpname
	slp := getSLP(t)
	if slp == nil {
		t.delTask()
		return
	}

	t.slp = slp
	t.process(jsip)
}
