// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// Timer

package rtclib

import (
	"time"
)

type Timer struct {
	timer   *time.Timer
	quit    chan bool
	handler func(d interface{})
	data    interface{}
}

func NewTimer(d time.Duration, f func(interface{}), p interface{}) *Timer {
	t := time.NewTimer(d)

	timer := &Timer{
		timer:   t,
		quit:    make(chan bool),
		handler: f,
		data:    p,
	}

	go func() {
		select {
		case <-t.C:
			timer.handler(timer.data)
		case <-timer.quit:
			t.Stop()
		}

		timer.timer = nil
	}()

	return timer
}

func (t *Timer) Stop() {
	if t.timer != nil {
		t.timer = nil
		t.quit <- true
	}
}

func (t *Timer) Reset(d time.Duration) {
	if t.timer != nil {
		t.timer.Reset(d)
	}
}
