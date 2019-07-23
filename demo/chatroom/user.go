// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// ChatRoom demo
// chatroom user

package main

import (
	"rtclib"
	"time"
)

type user struct {
	userid   string
	nickname string

	task *rtclib.Task
	room *room
	conf *config

	timer *time.Timer
	sub   chan uint64
	msgC  chan *rtclib.JSIP
	res   chan *rtclib.JSIP
	msgs  map[string]*rtclib.JSIP
}

func (r *room) newUser(userid string, nickname string, expire uint64) *user {
	u := &user{
		userid:   userid,
		nickname: nickname,

		task: r.task,
		room: r,
		conf: r.conf,

		timer: time.NewTimer(time.Duration(expire) * time.Second),
		sub:   make(chan uint64, r.conf.Qsize),
		msgC:  make(chan *rtclib.JSIP, r.conf.Qsize),
		res:   make(chan *rtclib.JSIP, r.conf.Qsize),
		msgs:  make(map[string]*rtclib.JSIP),
	}

	go u.loop()

	// TODO Send history message

	return u
}

func (u *user) process(msg *rtclib.JSIP) {
	u.msgC <- msg
}

func (u *user) subscribe(expire uint64) {
	u.sub <- expire
}

func (u *user) result(res *rtclib.JSIP) {
	if res.Type == rtclib.TERM {
		return
	}

	u.res <- res
}

func (u *user) loop() {
	defer func() {
		u.room.delUser(u.userid)
	}()

	for {
		select {
		case msg := <-u.msgC:
			dlg := u.task.NewDialogueIDWithEntry(u.result)

			u.msgs[dlg] = msg

			m := rtclib.JSIPMsgClone(msg, dlg)
			m.RequestURI = u.userid
			if len(m.Router) > 1 {
				m.Router = m.Router[1:]
			} else {
				m.Router = []string{}
			}
			rtclib.SendMsg(m)

		case res := <-u.res:
			dlg := res.DialogueID

			req := u.msgs[dlg]
			delete(u.msgs, dlg)

			if res.Code == 408 { // peer no response
				u.process(req)
			}

		case expire := <-u.sub:
			if expire == 0 {
				// User Unsubscribe
				return
			} else {
				u.timer.Reset(time.Duration(expire) * time.Second)
			}

		case <-u.timer.C:
			u.task.LogError("Wait for SUBSCRIBE for user(%s) timeout", u.userid)
			return
		}
	}
}
