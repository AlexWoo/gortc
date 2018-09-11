package main

import (
	"log"
	"rtclib"
	"time"

	"github.com/alexwoo/golib"
)

type user struct {
	userid   string
	nickname string
	timer    *golib.Timer
}

func (u *user) subTimeout(d interface{}) {
	r := d.(*room)

	r.quit <- u
}

type room struct {
	name string            // roomid
	msgs chan *msg         // user Subscriber or Message
	resp chan *rtclib.JSIP // resp for request broadcast
	quit chan *user        // user quit room

	task *rtclib.Task

	// key: userid, value: user, record users register in rooms
	users map[string]*user
}

func (r *room) newUser(userid string, nickname string,
	expire time.Duration) *user {

	u := &user{
		userid:   userid,
		nickname: nickname,
	}

	u.timer = golib.NewTimer(expire, u.subTimeout, r)

	return u
}

func (r *room) delUser(u *user) bool {
	ctx := r.task.GetCtx().(*ctx)

	u.timer.Stop()

	delete(r.users, u.userid)
	log.Printf("Delete user %s, %v", u.userid, r.users)
	if len(r.users) == 0 {
		ctx.roomdel <- r.name
		return true
	}

	return false
}

func newRoom(name string, qsize int, subTimeout time.Duration,
	task *rtclib.Task) *room {

	r := &room{
		name: name,
		msgs: make(chan *msg, qsize),
		resp: make(chan *rtclib.JSIP, 1),
		quit: make(chan *user, qsize),

		task: task,

		users: make(map[string]*user),
	}

	go r.process()

	return r
}

func (r *room) processSubscriber(sub *msg) {
	exp, _ := sub.req.GetInt("Expire")
	expire := time.Duration(exp) * time.Second
	userid, _ := sub.req.GetString("P-Asserted-Identity")

	user := r.users[userid]
	if user == nil {
		if expire > 0 { // User register in chatroom
			user = r.newUser(userid, sub.req.From, expire)
			r.users[userid] = user
			sub.res <- rtclib.JSIPMsgRes(sub.req, 200)
			log.Printf("User %s register in %s", userid, r.name)
		} else {
			sub.res <- rtclib.JSIPMsgRes(sub.req, 404)
			log.Printf("User %s not register in %s", userid, r.name)
		}

		return
	}

	if expire > 0 { // User refresh register state in chatroom
		user.timer.Reset(expire)
		sub.res <- rtclib.JSIPMsgRes(sub.req, 200)
	} else { // User deregister from chatroom
		r.quit <- user
		sub.res <- rtclib.JSIPMsgRes(sub.req, 200)
	}
}

func (r *room) processMessage(mess *msg) {
	ctx := r.task.GetCtx().(*ctx)

	userid, _ := mess.req.GetString("P-Asserted-Identity")

	user := r.users[userid]
	if user == nil {
		mess.res <- rtclib.JSIPMsgRes(mess.req, 404)
		log.Printf("User %s not register in %s when receive Message",
			userid, r.name)

		return
	}

	mess.res <- rtclib.JSIPMsgRes(mess.req, 200)

	for id := range r.users {
		if id == userid {
			continue
		}

		message := rtclib.JSIPMsgClone(mess.req, r.task.NewDialogueID())
		message.RequestURI = id

		if len(message.Router) > 0 {
			message.Router = message.Router[1:]
		}

		message.SetString("P-Asserted-Identity", rtclib.Realm())

		req := &req{
			req:  message,
			room: r,
		}

		ctx.reqs <- req
	}
}

func (r *room) process() {
	for {
		select {
		case u := <-r.quit:
			if r.delUser(u) {
				return
			}

		case resp := <-r.resp:
			log.Printf("Receive response: %s", resp.Abstract())

		case msg := <-r.msgs:
			switch msg.req.Type {
			case rtclib.SUBSCRIBE:
				r.processSubscriber(msg)
			case rtclib.MESSAGE:
				r.processMessage(msg)
			}
		}
	}
}
