// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// RTC Task manager

package rtclib

import (
	uuid "github.com/satori/go.uuid"
)

const (
	CONTINUE = iota
	FINISH
)

type SLP interface {
	Process(jsip *JSIP) int
}

type Task struct {
	Name string
	SLP  SLP
	dlgs []string
}

var tasks = make(map[string]*Task)

func (t *Task) NewDialogueID() string {
	u1, _ := uuid.NewV1()
	dlg := realm + u1.String()

	t.dlgs = append(t.dlgs, dlg)
	tasks[dlg] = t

	return dlg
}

func (t *Task) DelTask() {
	for _, dlg := range t.dlgs {
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
	t.dlgs = append(t.dlgs, dlg)
	tasks[dlg] = t

	return t
}
