package main

import (
	"rtclib"
	"sync"
	"time"

	"github.com/alexwoo/golib"
)

type msg struct {
	req *rtclib.JSIP
	res chan *rtclib.JSIP
}

type ctx struct {
	msgs    chan *msg   // msg send to room
	roomdel chan string // room name wait for delete

	// key: room name, value: room, for room manager
	rooms     map[string]*room
	roomsLock sync.RWMutex
}

type config struct {
	Qsize     golib.Size    `default:"1k"`
	Timeout   time.Duration `default:"5s"`
	Apiserver string
}

type ChatRoom struct {
	task    *rtclib.Task
	qsize   int
	timeout time.Duration
}

var (
	c    *ctx
	conf *config
)

func loadConfig() error {
	file := rtclib.FullPath("conf/chatroom/chatroom.conf")
	pconf := &config{}

	err := golib.JsonConfigFile(file, pconf)
	if err == nil {
		conf = pconf
	}

	return err
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	if conf == nil {
		err := loadConfig()
		if err != nil {
			task.LogError("Load Config error: %s", err)
			return nil
		}
	}

	slp := &ChatRoom{
		task:    task,
		qsize:   int(conf.Qsize),
		timeout: conf.Timeout,
	}

	return slp
}

func (slp *ChatRoom) NewSLPCtx() interface{} {
	if c != nil {
		return c
	}

	c = &ctx{
		msgs:    make(chan *msg, slp.qsize),
		roomdel: make(chan string),
		rooms:   make(map[string]*room),
	}

	return c
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
		slp.task.LogError("Process msg %s expire\n", jsip.String())
		rtclib.SendMsg(resp)
	}

	slp.task.SetFinished()
}

func (slp *ChatRoom) OnLoad(jsip *rtclib.JSIP) {
	if jsip == nil {
		go slp.roomManager()

		return
	}
}

func (slp *ChatRoom) onMsg(m *msg) {
	ctx := slp.task.GetCtx().(*ctx)

	var expire int64
	if m.req.Type == rtclib.SUBSCRIBE {
		expr, ok := m.req.GetInt("Expire")
		if !ok {
			m.res <- rtclib.JSIPMsgRes(m.req, 400)
			slp.task.LogError("Receive SUBSCRIBE but no expire")
			return
		}
		expire = expr
	}

	_, ok := m.req.GetString("P-Asserted-Identity")
	if !ok {
		m.res <- rtclib.JSIPMsgRes(m.req, 400)
		slp.task.LogError("Receive %s but no P-Asserted-Identity", m.req.Name())
		return
	}

	roomid := m.req.To
	ctx.roomsLock.RLock()
	room := ctx.rooms[roomid]
	ctx.roomsLock.RUnlock()

	if room != nil {
		room.msgs <- m
		return
	}

	if m.req.Type == rtclib.SUBSCRIBE && expire > 0 {
		room = newRoom(roomid, slp.qsize, time.Duration(expire)*time.Second,
			slp.task)
		ctx.roomsLock.Lock()
		ctx.rooms[roomid] = room
		ctx.roomsLock.Unlock()
		room.msgs <- m
		return
	}

	resp := rtclib.JSIPMsgRes(m.req, 404)
	m.res <- resp
	slp.task.LogError("Receive msg %s, but no room found", m.req.String())
}

func (slp *ChatRoom) roomManager() {
	ctx := slp.task.GetCtx().(*ctx)

	for {
		select {
		case msg := <-ctx.msgs:
			// Subscriber or Messeges From User
			slp.onMsg(msg)

		case name := <-ctx.roomdel:
			// room need to deleted
			ctx.roomsLock.Lock()
			delete(ctx.rooms, name)
			slp.task.LogError("All users quit room %s %v\n", name, ctx.rooms)
			ctx.roomsLock.Unlock()
		}
	}
}
