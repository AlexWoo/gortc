// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//
// rtclib timer

package rtclib

import (
	"time"
)

// go lib timer struct, use NewTimer to create.
// handler will execute after 1 second in example below
//
//	func handler(p interface{}) {
//		i := p.(int)
//		fmt.Println(i)
//	}
//
//	timer := golib.NewTimer(1 * time.Second, handler, 10)
type Timer struct {
	timer   *time.Timer
	quit    chan bool
	handler func(d interface{})
	data    interface{}
}

func (timer *Timer) startTimer(d time.Duration) {
	timer.timer = time.NewTimer(d)

	go func() {
		select {
		case <-timer.timer.C:
			if timer.handler != nil {
				timer.handler(timer.data)
			}
		case <-timer.quit:
			return
		}
	}()
}

// NewTimer creates a new Timer,
// that will call function f with paras p after duration d
func NewTimer(d time.Duration, f func(interface{}), p interface{}) *Timer {
	timer := &Timer{
		quit:    make(chan bool),
		handler: f,
		data:    p,
	}

	timer.startTimer(d)

	return timer
}

// Stop prevents the Timer From firing
func (t *Timer) Stop() {
	if t.timer == nil {
		return
	}

	if t.timer.Stop() {
		// timer had been active
		t.quit <- true
	}
}

// Reset changes the timer to expire after duration d,
// if timer had expired or been stoped, it will restart timer again
func (t *Timer) Reset(d time.Duration) {
	if t.timer == nil {
		return
	}

	if !t.timer.Reset(d) {
		// timer had expired or been stopped
		t.startTimer(d)
	}
}
