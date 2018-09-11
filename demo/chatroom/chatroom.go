package main

import (
	"log"
	"rtclib"
	"time"
)

type msg struct {
	req *rtclib.JSIP
	res chan *rtclib.JSIP
}

type req struct {
	req  *rtclib.JSIP
	room *room
}

type ctx struct {
	msgs    chan *msg         // msg send to room
	reqs    chan *req         // req send outside
	resp    chan *rtclib.JSIP // resp for req
	roomdel chan string       // room name wait for delete

	// key: room name, value: room, for room manager
	rooms map[string]*room

	// key: DialogueID, value: room, for resp send to room
	dlgs map[string]*room
}

type ChatRoom struct {
	task    *rtclib.Task
	qsize   int
	timeout time.Duration
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	slp := &ChatRoom{
		task:    task,
		qsize:   4096,
		timeout: 5 * time.Second,
	}

	return slp
}

func (slp *ChatRoom) NewSLPCtx() interface{} {
	ctx := &ctx{
		msgs:    make(chan *msg, slp.qsize),
		reqs:    make(chan *req, slp.qsize),
		resp:    make(chan *rtclib.JSIP),
		roomdel: make(chan string),
		rooms:   make(map[string]*room),
		dlgs:    make(map[string]*room),
	}

	return ctx
}

func (slp *ChatRoom) Process(jsip *rtclib.JSIP) {
	ctx := slp.task.GetCtx().(*ctx)

	if jsip.Type == rtclib.CANCEL {
		slp.task.SetFinished()
		return
	}

	m := &msg{
		req: jsip,
		res: make(chan *rtclib.JSIP, 1),
	}

	ctx.msgs <- m

	timer := time.NewTimer(slp.timeout)

	select {
	case resp := <-m.res:
		rtclib.SendMsg(resp)
		timer.Stop()
	case <-timer.C:
		resp := rtclib.JSIPMsgRes(jsip, 408)
		log.Printf("Process msg %s expire\n", jsip.Abstract())
		rtclib.SendMsg(resp)
	}

	slp.task.SetFinished()
}

func (slp *ChatRoom) OnLoad(jsip *rtclib.JSIP) {
	ctx := slp.task.GetCtx().(*ctx)

	if jsip == nil {
		go slp.roomManager()

		return
	}

	ctx.resp <- jsip
}

func (slp *ChatRoom) onMsg(m *msg) {
	ctx := slp.task.GetCtx().(*ctx)

	var expire int64
	if m.req.Type == rtclib.SUBSCRIBE {
		expr, ok := m.req.GetInt("Expire")
		if !ok {
			m.res <- rtclib.JSIPMsgRes(m.req, 400)
			log.Printf("Receive SUBSCRIBE but no expire")
			return
		}
		expire = expr
	}

	_, ok := m.req.GetString("P-Asserted-Identity")
	if !ok {
		m.res <- rtclib.JSIPMsgRes(m.req, 400)
		log.Printf("Receive %s but no P-Asserted-Identity", m.req.Name())
		return
	}

	roomid := m.req.To
	room := ctx.rooms[roomid]

	if room != nil {
		room.msgs <- m
		return
	}

	if m.req.Type == rtclib.SUBSCRIBE && expire > 0 {
		room = newRoom(roomid, slp.qsize, time.Duration(expire)*time.Second,
			slp.task)
		ctx.rooms[roomid] = room
		room.msgs <- m
		return
	}

	resp := rtclib.JSIPMsgRes(m.req, 404)
	m.res <- resp
	log.Printf("Receive msg %s, but no room found", m.req.Abstract())
}

func (slp *ChatRoom) roomManager() {
	ctx := slp.task.GetCtx().(*ctx)

	for {
		select {
		case msg := <-ctx.msgs:
			// Subscriber or Messeges From User
			slp.onMsg(msg)

		case req := <-ctx.reqs:
			// Request Send to outside
			dlg := req.req.DialogueID
			ctx.dlgs[dlg] = req.room
			rtclib.SendMsg(req.req)

		case resp := <-ctx.resp:
			// Response Receive from outside
			dlg := resp.DialogueID
			room := ctx.dlgs[dlg]
			room.resp <- resp

			delete(ctx.dlgs, dlg)

		case name := <-ctx.roomdel:
			// room need to deleted
			delete(ctx.rooms, name)
			log.Printf("All users quit room %s %v\n", name, ctx.rooms)
		}
	}
}
