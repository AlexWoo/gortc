// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// ChatRoom demo
// chatroom room

package main

import (
	"rtclib"
	"sync"
)

type room struct {
	roomid string

	task        *rtclib.Task
	roomManager *roomManager
	conf        *config

	msgC chan *rtclib.JSIP

	users     map[string]*user
	userdel   chan string
	usersLock sync.RWMutex // Lock used in room goroutine and API goroutine
}

func (m *roomManager) newRoom(roomid string) *room {
	r := &room{
		roomid: roomid,

		task:        m.task,
		roomManager: m,
		conf:        m.conf,

		msgC: make(chan *rtclib.JSIP, m.conf.Qsize),

		users:   make(map[string]*user),
		userdel: make(chan string, m.conf.Qsize),
	}

	go r.loop()

	return r
}

func (r *room) process(msg *rtclib.JSIP) {
	r.msgC <- msg
}

func (r *room) delUser(userid string) {
	r.userdel <- userid
}

func (r *room) loop() {
	defer func() {
		r.roomManager.delRoom(r.roomid)
	}()

	for {
		select {
		case msg := <-r.msgC:
			if msg.Type == rtclib.SUBSCRIBE {
				r.processSubscribe(msg)
			} else {
				r.processMessage(msg)
			}

		case name := <-r.userdel:
			r.usersLock.Lock()
			delete(r.users, name)

			if len(r.users) == 0 {
				// All user quit from room
				r.usersLock.Unlock()
				return
			}

			r.usersLock.Unlock()
		}
	}
}

func (r *room) processSubscribe(msg *rtclib.JSIP) {
	userid, _ := msg.GetString("P-Asserted-Identity")
	r.usersLock.RLock()
	user := r.users[userid]
	r.usersLock.RUnlock()

	expire, _ := msg.GetUint("Expire")

	if expire == 0 {
		if user == nil {
			res := rtclib.JSIPMsgRes(msg, 404)
			res.SetString("Reason", "User Not Exist")
			rtclib.SendMsg(res)

			r.task.LogError("Recv SUBSCRIBE(0) but user(%s) not in room, %s", userid, msg.Suffix())

			return
		}

		user.subscribe(0)
	} else {
		if user == nil {
			user = r.newUser(userid, msg.From, expire)

			r.usersLock.Lock()
			r.users[userid] = user
			r.usersLock.Unlock()
		} else {
			user.subscribe(expire)
		}
	}

	rtclib.SendMsg(rtclib.JSIPMsgRes(msg, 200))
}

func (r *room) processMessage(msg *rtclib.JSIP) {
	userid, _ := msg.GetString("P-Asserted-Identity")
	r.usersLock.RLock()
	user := r.users[userid]
	r.usersLock.RUnlock()

	if user == nil {
		res := rtclib.JSIPMsgRes(msg, 404)
		res.SetString("Reason", "User Not Exist")
		rtclib.SendMsg(res)

		r.task.LogError("Recv MESSAGE but user(%s) not in room, %s", userid, msg.Suffix())

		return
	}

	rtclib.SendMsg(rtclib.JSIPMsgRes(msg, 200))

	// TODO save message in storage

	// broadcast message to all subscriber
	r.usersLock.RLock()
	defer r.usersLock.RUnlock()

	for id, u := range r.users {
		if id == userid {
			continue
		}

		u.process(msg)
	}
}
