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

func (timer *Timer) startTimer(d time.Duration) {
	t := time.NewTimer(d)

	go func() {
		select {
		case <-t.C:
			if timer.handler != nil {
				timer.handler(timer.data)
			}
		case <-timer.quit:
			t.Stop()
		}

		timer.timer = nil
	}()

	timer.timer = t
}

func NewTimer(d time.Duration, f func(interface{}), p interface{}) *Timer {
	timer := &Timer{
		quit:    make(chan bool),
		handler: f,
		data:    p,
	}

	timer.startTimer(d)

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
	} else {
		t.startTimer(d)
	}
}
