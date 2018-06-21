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
	timer := &Timer{
		timer:   time.NewTimer(d),
		quit:    make(chan bool),
		handler: f,
		data:    p,
	}

	go func() {
		select {
		case <-timer.timer.C:
			timer.handler(timer.data)
		case <-timer.quit:
			timer.timer.Stop()
		}

		timer.timer = nil
	}()

	return timer
}

func (t *Timer) Stop() {
	if t.timer != nil {
		t.quit <- true
	}
}

func (t *Timer) Reset(d time.Duration) {
	t.timer.Reset(d)
}
