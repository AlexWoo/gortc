// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// ChatRoom demo
// chatroom manager

package main

import (
	"rtclib"
	"sync"

	"github.com/alexwoo/golib"
)

type config struct {
	Qsize golib.Size `default:"1k"`
}

type roomManager struct {
	task *rtclib.Task
	conf *config

	msgC chan *rtclib.JSIP

	rooms     map[string]*room
	roomdel   chan string
	roomsLock sync.RWMutex // Lock used in roomManager goroutine and API goroutine
}

func (slp *ChatRoom) newRoomManager() *roomManager {
	m := &roomManager{
		task:  slp.task,
		rooms: make(map[string]*room),
	}

	err := m.loadConfig()
	if err != nil {
		return nil
	}

	m.msgC = make(chan *rtclib.JSIP, m.conf.Qsize)
	m.roomdel = make(chan string, m.conf.Qsize)

	go m.loop()

	return m
}

func (m *roomManager) loadConfig() error {
	file := rtclib.FullPath("conf/chatroom/chatroom.conf")
	pconf := &config{}

	err := golib.JsonConfigFile(file, pconf)
	if err == nil {
		m.conf = pconf
	} else {
		m.task.LogError("Load Config error: %s", err)
	}

	return err
}

func (m *roomManager) process(msg *rtclib.JSIP) {
	m.msgC <- msg
}

func (m *roomManager) delRoom(roomid string) {
	m.roomdel <- roomid
}

func (m *roomManager) loop() {
	for {
		select {
		case msg := <-m.msgC:
			if msg.Type == rtclib.SUBSCRIBE && msg.Code == 0 {
				m.processSubscribe(msg)
			} else if msg.Type == rtclib.MESSAGE && msg.Code == 0 {
				m.processMessage(msg)
			} else {
				if msg.Code == 0 {
					res := rtclib.JSIPMsgRes(msg, 400)
					res.SetString("Reason", "Unexpected msg")
					rtclib.SendMsg(res)
				}

				m.task.LogError("Process unexpected msg in roomManager %s", msg.String())
			}

		case name := <-m.roomdel:
			m.roomsLock.Lock()
			delete(m.rooms, name)
			m.roomsLock.Unlock()
		}
	}
}

func (m *roomManager) processSubscribe(msg *rtclib.JSIP) {
	expr, ok := msg.GetUint("Expire")
	if !ok {
		res := rtclib.JSIPMsgRes(msg, 400)
		res.SetString("Reason", "No Expire")
		rtclib.SendMsg(res)

		m.task.LogError("Recv SUBSCRIBE but no Expire, %s", msg.Suffix())

		return
	}

	expire := expr

	if _, ok := msg.GetString("P-Asserted-Identity"); !ok {
		res := rtclib.JSIPMsgRes(msg, 400)
		res.SetString("Reason", "No P-Asserted-Identity")
		rtclib.SendMsg(res)

		m.task.LogError("Recv SUBSCRIBE but no P-Asserted-Identity, %s", msg.Suffix())

		return
	}

	roomid := msg.To
	m.roomsLock.RLock()
	room := m.rooms[roomid]
	m.roomsLock.RUnlock()

	if room == nil {
		if expire == 0 {
			res := rtclib.JSIPMsgRes(msg, 404)
			res.SetString("Reason", "Room Not Exist")
			rtclib.SendMsg(res)

			m.task.LogError("Recv SUBSCRIBE(0) but room(%s) not exist, %s", roomid, msg.Suffix())

			return
		}

		room = m.newRoom(roomid)

		m.roomsLock.Lock()
		m.rooms[roomid] = room
		m.roomsLock.Unlock()
	}

	room.process(msg)
}

func (m *roomManager) processMessage(msg *rtclib.JSIP) {
	if _, ok := msg.GetString("P-Asserted-Identity"); !ok {
		res := rtclib.JSIPMsgRes(msg, 400)
		res.SetString("Reason", "No P-Asserted-Identity")
		rtclib.SendMsg(res)

		m.task.LogError("Recv MESSAGE but no P-Asserted-Identity, %s", msg.Suffix())

		return
	}

	roomid := msg.To
	m.roomsLock.RLock()
	room := m.rooms[roomid]
	m.roomsLock.RUnlock()

	if room == nil {
		res := rtclib.JSIPMsgRes(msg, 404)
		res.SetString("Reason", "Room not exist")
		rtclib.SendMsg(res)

		m.task.LogError("Recv MESSAGE but room(%s) not exist, %s", roomid, msg.Suffix())

		return
	}

	room.process(msg)
}
