// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Session Test Case

package rtclib

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexwoo/golib"
)

var slog = golib.NewLog("session.log")

func TestSessionCommon(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestSessionCommon")

	assert(jsipSessionState(100).String() == "Unknown")
	assert(jsipSessionState(-100).String() == "Unknown")
	assert(jsipSessionState(0).String() == "Unknown")
	assert(INVITE_18X.String() == "INVITE_18X")

	assert(inviteSession(JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(ACK, "jsip.com", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(BYE, "jsip.co)m", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(CANCEL, "jsip.com", "jsip", "jsip", "123456")))
	assert(!inviteSession(JSIPMsgReq(REGISTER, "jsip.com", "jsip", "jsip", "123456")))
	assert(!inviteSession(JSIPMsgReq(OPTIONS, "jsip.com", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(INFO, "jsip.com", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(UPDATE, "jsip.com", "jsip", "jsip", "123456")))
	assert(inviteSession(JSIPMsgReq(PRACK, "jsip.com", "jsip", "jsip", "123456")))
	assert(!inviteSession(JSIPMsgReq(SUBSCRIBE, "jsip.com", "jsip", "jsip", "123456")))
	assert(!inviteSession(JSIPMsgReq(MESSAGE, "jsip.com", "jsip", "jsip", "123456")))
	assert(!inviteSession(JSIPMsgReq(NOTIFY, "jsip.com", "jsip", "jsip", "123456")))
}

func TestGetProcess(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestGetProcess")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	m.CSeq = 1

	init := &jsipSessionInit{
		sessionFailureCount: 3,
		sessionTimer:        time.Second * 3,
		prTimer:             time.Second * 1,
		qsize:               1024,
		msg:                 make(chan *JSIP),
		term:                make(chan string),
	}

	s := createSession(m, init, slog)
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sInit))

	s.state = INVITE_18X
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.s18X))

	s.state = INVITE_PRACK
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sPrack))

	s.state = INVITE_UPDATE
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sUpdate))

	s.state = INVITE_200
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.s200))

	s.state = INVITE_ACK
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sAck))

	s.state = INVITE_REINV
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sReInv))

	s.state = INVITE_RE200
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sRe200))

	s.state = INVITE_ERR
	assert(fmt.Sprintf("%p", s.getProcess()) == fmt.Sprintf("%p", s.sErr))

	s.state = INVITE_END
	assert(s.getProcess() == nil)
}

func testSessionMsg(ct []check, init *jsipSessionInit) {
	for _, c := range ct {
		msg := <-init.msg
		if msg.recv {
			fmt.Println("	recv msg:", msg.String())
		} else {
			fmt.Println("	send msg:", msg.String())
		}
		assert(msg.Type == c.typ)
		assert(msg.Code == c.code)
		assert(msg.recv == c.recv)
	}
}

func TestProcessSession(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestProcessSession")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")

	init := &jsipSessionInit{
		sessionFailureCount: 3,
		sessionTimer:        time.Second * 10,
		prTimer:             time.Second * 3,
		transTimer:          time.Second * 1,
		qsize:               1024,
		msg:                 make(chan *JSIP, 1024),
		term:                make(chan string, 1),
	}

	s := createSession(m, init, slog)
	msg := <-init.msg
	assert(m == msg)

	update := JSIPMsgUpdate(m)
	update180 := JSIPMsgRes(update, 180)
	update200 := JSIPMsgRes(update, 200)
	update408 := JSIPMsgRes(update, 408)

	resp100 := JSIPMsgRes(m, 100)
	resp180 := JSIPMsgRes(m, 180)
	resp202 := JSIPMsgRes(m, 202)
	resp200 := JSIPMsgRes(m, 200)
	resp404 := JSIPMsgRes(m, 404)

	cancel := JSIPMsgCancel(m)

	prack := JSIPMsgReq(PRACK, m.RequestURI, m.From, m.To, m.DialogueID)
	prack180 := JSIPMsgRes(prack, 180)
	prack200 := JSIPMsgRes(prack, 200)
	prack487 := JSIPMsgRes(prack, 487)

	ack := JSIPMsgAck(resp200)

	bye := JSIPMsgBye(m)

	reinvite := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	reresp180 := JSIPMsgRes(reinvite, 180)
	reresp200 := JSIPMsgRes(reinvite, 200)
	reresp408 := JSIPMsgRes(reinvite, 408)

	fmt.Println("++++++++++processSessionUpdate")

	update.recv = false
	state := s.state
	st, err := s.processSessionUpdate(update)
	assert(st == state)
	assert(err.Error() == "Send session update")

	m.recv = false
	update.recv = true
	state = s.state
	st, err = s.processSessionUpdate(update)
	assert(st == state)
	assert(err.Error() == "Session update direction")

	m.recv = true
	update.recv = true
	state = s.state
	st, err = s.processSessionUpdate(update)
	assert(st == state)
	assert(err.Error() == "Ignore")
	ct := []check{
		check{typ: UPDATE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	fmt.Println("++++++++++processSessionUpdateResp")

	update200.recv = false
	state = s.state
	st, err = s.processSessionUpdateResp(update200)
	assert(st == state)
	assert(err.Error() == "Send session update response")

	m.recv = true
	update200.recv = true
	state = s.state
	st, err = s.processSessionUpdateResp(update200)
	assert(st == state)
	assert(err.Error() == "Session update response direction")

	m.recv = false
	update200.recv = true
	state = s.state
	st, err = s.processSessionUpdateResp(update200)
	assert(st == state)
	assert(err.Error() == "Ignore")

	m.recv = false
	update180.recv = true
	state = s.state
	st, err = s.processSessionUpdateResp(update180)
	assert(st == state)
	assert(err.Error() == "Session update unexpected response")

	m.recv = false
	update408.recv = true
	state = s.state
	st, err = s.processSessionUpdateResp(update408)
	assert(st == state)
	assert(err.Error() == "Session update unexpected response")

	fmt.Println("++++++++++process1XX")

	m.recv = true
	resp180.recv = true
	state = s.state
	st, err = s.process1XX(resp180)
	assert(st == state)
	assert(err.Error() == "Response direction")

	s.cancelled = true

	m.recv = true
	resp180.recv = false
	state = s.state
	st, err = s.process1XX(resp180)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = true
	resp100.recv = false
	state = s.state
	st, err = s.process1XX(resp100)
	assert(st == state)
	assert(err.Error() == "Ignore")

	m.recv = true
	resp180.recv = false
	state = s.state
	st, err = s.process1XX(resp180)
	assert(st == INVITE_18X)
	assert(err == nil)

	fmt.Println("++++++++++process2XX")

	m.recv = true
	resp200.recv = true
	state = s.state
	st, err = s.process2XX(resp200)
	assert(st == state)
	assert(err.Error() == "Response direction")

	m.recv = true
	resp202.recv = false
	state = s.state
	st, err = s.process2XX(resp202)
	assert(st == INVITE_END)
	assert(err.Error() == "Unexpected response")
	ct = []check{
		check{typ: BYE, code: 0, recv: true},
	}
	testSessionMsg(ct, init)

	s.cancelled = true

	m.recv = false
	resp200.recv = true
	state = s.state
	st, err = s.process2XX(resp200)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = false
	resp200.recv = true
	st, err = s.process2XX(resp200)
	assert(st == INVITE_200)
	assert(err == nil)
	assert(s.sessionTimeout == init.sessionTimer/3)

	m.recv = true
	resp200.recv = false
	st, err = s.process2XX(resp200)
	assert(st == INVITE_200)
	assert(err == nil)
	assert(s.sessionTimeout == init.sessionTimer)

	fmt.Println("++++++++++processErrResp")

	m.recv = true
	resp404.recv = true
	state = s.state
	st, err = s.processErrResp(resp404)
	assert(st == state)
	assert(err.Error() == "Response direction")

	m.recv = true
	resp404.recv = false
	st, err = s.processErrResp(resp404)
	assert(st == INVITE_ERR)
	assert(err == nil)

	m.recv = false
	resp404.recv = true
	st, err = s.processErrResp(resp404)
	assert(st == INVITE_END)
	assert(err == nil)
	ct = []check{
		check{typ: ACK, code: 0, recv: false},
	}
	testSessionMsg(ct, init)

	fmt.Println("++++++++++processErrAck")

	m.recv = true
	ack.recv = false
	state = s.state
	st, err = s.processErrAck(ack)
	assert(st == state)
	assert(err.Error() == "ACK direction")

	m.recv = true
	ack.recv = true
	st, err = s.processErrAck(ack)
	assert(st == INVITE_END)
	assert(err.Error() == "Ignore")

	fmt.Println("++++++++++processCancel")

	m.recv = true
	cancel.recv = false
	state = s.state
	st, err = s.processCancel(cancel)
	assert(st == state)
	assert(err.Error() == "CANCEL direction")

	s.cancelled = true

	m.recv = true
	cancel.recv = true
	state = s.state
	st, err = s.processCancel(cancel)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = true
	cancel.recv = true
	st, err = s.processCancel(cancel)
	assert(st == INVITE_ERR)
	assert(err == nil)
	ct = []check{
		check{typ: CANCEL, code: 200, recv: false},
		check{typ: INVITE, code: 487, recv: false},
	}
	testSessionMsg(ct, init)

	m.recv = false
	cancel.recv = false
	state = s.state
	st, err = s.processCancel(cancel)
	assert(st == state)
	assert(err == nil)
	assert(s.cancelled)

	s.cancelled = false

	fmt.Println("++++++++++processPrack")

	m.recv = true
	prack.recv = false
	state = s.state
	st, err = s.processPrack(prack)
	assert(st == state)
	assert(err.Error() == "PRACK direction")

	s.cancelled = true

	m.recv = false
	prack.recv = false
	state = s.state
	st, err = s.processPrack(prack)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = false
	prack.recv = false
	st, err = s.processPrack(prack)
	assert(st == INVITE_PRACK)
	assert(err == nil)

	fmt.Println("++++++++++processPrackResp")

	m.recv = false
	prack200.recv = false
	state = s.state
	st, err = s.processPrackResp(prack200)
	assert(st == state)
	assert(err.Error() == "PRACK response direction")

	s.cancelled = true

	m.recv = true
	prack200.recv = false
	state = s.state
	st, err = s.processPrackResp(prack200)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = true
	prack180.recv = false
	state = s.state
	st, err = s.processPrackResp(prack180)
	assert(st == state)
	assert(err.Error() == "PRACK provisional response")

	m.recv = true
	prack200.recv = false
	st, err = s.processPrackResp(prack200)
	assert(st == INVITE_18X)
	assert(err == nil)

	m.recv = true
	prack487.recv = false
	st, err = s.processPrackResp(prack487)
	assert(st == INVITE_18X)
	assert(err.Error() == "PRACK error response")

	fmt.Println("++++++++++processUpdate")

	s.cancelled = true

	m.recv = true
	update.recv = false
	state = s.state
	st, err = s.processUpdate(update)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	m.recv = true
	update.recv = false
	st, err = s.processUpdate(update)
	assert(s.updateRecv == update.recv)
	assert(st == INVITE_UPDATE)
	assert(err == nil)

	fmt.Println("++++++++++processUpdateResp")

	s.updateRecv = false
	update200.recv = false
	state = s.state
	st, err = s.processUpdateResp(update200)
	assert(st == state)
	assert(err.Error() == "UPDATE response direction")

	s.cancelled = true

	s.updateRecv = true
	update200.recv = false
	state = s.state
	st, err = s.processUpdateResp(update200)
	assert(st == state)
	assert(err.Error() == "Ignore")

	s.cancelled = false

	s.updateRecv = true
	update180.recv = false
	state = s.state
	st, err = s.processUpdateResp(update180)
	assert(st == state)
	assert(err.Error() == "UPDATE provisional response")

	s.updateRecv = true
	update200.recv = false
	st, err = s.processUpdateResp(update200)
	assert(st == INVITE_18X)
	assert(err == nil)

	s.updateRecv = true
	update408.recv = false
	st, err = s.processUpdateResp(update408)
	assert(st == INVITE_18X)
	assert(err.Error() == "UPDATE error response")

	fmt.Println("++++++++++processAck")

	s.inviteRecv = true
	ack.recv = false
	state = s.state
	st, err = s.processAck(ack)
	assert(st == state)
	assert(err.Error() == "ACK direction")

	s.inviteRecv = false
	ack.recv = false
	st, err = s.processAck(update200)
	assert(st == INVITE_ACK)
	assert(err == nil)

	fmt.Println("++++++++++processBye")

	bye.recv = true
	st, err = s.processBye(bye)
	assert(st == INVITE_END)
	assert(err == nil)
	ct = []check{
		check{typ: BYE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	bye.recv = false
	st, err = s.processBye(bye)
	assert(st == INVITE_END)
	assert(err == nil)

	fmt.Println("++++++++++processReInvite")

	reinvite.recv = true
	st, err = s.processReInvite(reinvite)
	assert(s.inviteRecv == reinvite.recv)
	assert(st == INVITE_REINV)
	assert(err == nil)

	fmt.Println("++++++++++processReResp")

	s.inviteRecv = false
	reresp200.recv = false
	state = s.state
	st, err = s.processReResp(reresp200)
	assert(st == state)
	assert(err.Error() == "Re-INVITE response direction")

	s.inviteRecv = true
	reresp180.recv = false
	state = s.state
	st, err = s.processReResp(reresp180)
	assert(st == state)
	assert(err.Error() == "Re-INVITE provisional response")

	s.inviteRecv = true
	reresp200.recv = false
	st, err = s.processReResp(reresp200)
	assert(st == INVITE_RE200)
	assert(err == nil)

	s.inviteRecv = true
	reresp408.recv = false
	st, err = s.processReResp(reresp408)
	assert(st == INVITE_RE200)
	assert(err == nil)

	s.inviteRecv = false
	reresp408.recv = true
	st, err = s.processReResp(reresp408)
	assert(st == INVITE_ACK)
	assert(err == nil)
	ct = []check{
		check{typ: ACK, code: 0, recv: false},
	}
	testSessionMsg(ct, init)
}

func TestProcessSessionState(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestProcessSessionState")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	m.CSeq = 1

	init := &jsipSessionInit{
		sessionFailureCount: 3,
		sessionTimer:        time.Second * 10,
		prTimer:             time.Second * 3,
		transTimer:          time.Second * 1,
		qsize:               1024,
		msg:                 make(chan *JSIP, 1024),
		term:                make(chan string, 1),
	}

	s := createSession(m, init, slog)
	msg := <-init.msg
	assert(m == msg)

	resp100 := JSIPMsgRes(m, 100)
	resp180 := JSIPMsgRes(m, 180)
	resp200 := JSIPMsgRes(m, 200)
	resp404 := JSIPMsgRes(m, 404)

	cancel := JSIPMsgCancel(m)

	prack := JSIPMsgReq(PRACK, m.RequestURI, m.From, m.To, m.DialogueID)
	prack200 := JSIPMsgRes(prack, 200)

	update := JSIPMsgUpdate(m)
	update200 := JSIPMsgRes(update, 200)

	ack := JSIPMsgAck(resp200)

	reinvite := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	reinvite200 := JSIPMsgRes(reinvite, 200)

	bye := JSIPMsgBye(m)

	fmt.Println("++++++++++sInit")

	s.state = INVITE_INIT
	st, err := s.sInit(bye)
	assert(st == INVITE_INIT)
	assert(err.Error() == "Unexpected msg")

	s.state = INVITE_INIT
	st, err = s.sInit(prack)
	assert(st == INVITE_INIT)
	assert(err.Error() == "Unexpected msg")

	s.state = INVITE_INIT
	m.recv = false
	cancel.recv = false
	st, err = s.sInit(cancel)
	assert(st == INVITE_INIT)
	assert(err == nil)
	assert(s.cancelled)

	s.cancelled = false

	s.state = INVITE_INIT
	m.recv = false
	resp100.recv = true
	st, err = s.sInit(resp100)
	assert(st == INVITE_INIT)
	assert(err.Error() == "Ignore")

	s.state = INVITE_INIT
	m.recv = false
	resp180.recv = true
	st, err = s.sInit(resp180)
	assert(st == INVITE_18X)
	assert(err == nil)

	s.state = INVITE_INIT
	m.recv = false
	resp200.recv = true
	st, err = s.sInit(resp200)
	assert(st == INVITE_200)
	assert(err == nil)

	s.state = INVITE_INIT
	m.recv = false
	resp404.recv = true
	st, err = s.sInit(resp404)
	assert(st == INVITE_END)
	assert(err == nil)
	ct := []check{
		check{typ: ACK, code: 0, recv: false},
	}
	testSessionMsg(ct, init)

	fmt.Println("++++++++++s18X")

	s.state = INVITE_18X
	m.recv = true
	resp180.recv = false
	st, err = s.s18X(resp180)
	assert(st == INVITE_18X)
	assert(err == nil)

	s.state = INVITE_18X
	m.recv = true
	resp200.recv = false
	st, err = s.s18X(resp200)
	assert(st == INVITE_200)
	assert(err == nil)

	s.state = INVITE_18X
	m.recv = true
	resp404.recv = false
	st, err = s.s18X(resp404)
	assert(st == INVITE_ERR)
	assert(err == nil)

	s.state = INVITE_18X
	m.recv = true
	cancel.recv = true
	st, err = s.s18X(cancel)
	assert(st == INVITE_ERR)
	assert(err == nil)
	ct = []check{
		check{typ: CANCEL, code: 200, recv: false},
		check{typ: INVITE, code: 487, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_18X
	m.recv = true
	update.recv = false
	st, err = s.s18X(update)
	assert(st == INVITE_UPDATE)
	assert(err == nil)
	assert(!s.updateRecv)

	s.state = INVITE_18X
	m.recv = true
	prack.recv = true
	st, err = s.s18X(prack)
	assert(st == INVITE_PRACK)
	assert(err == nil)

	s.state = INVITE_18X
	st, err = s.s18X(ack)
	assert(st == INVITE_18X)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sPrack")

	s.state = INVITE_PRACK
	m.recv = false
	cancel.recv = false
	st, err = s.sPrack(cancel)
	assert(st == INVITE_PRACK)
	assert(err == nil)

	s.cancelled = false

	s.state = INVITE_PRACK
	m.recv = false
	prack200.recv = true
	st, err = s.sPrack(prack200)
	assert(st == INVITE_18X)
	assert(err == nil)

	s.state = INVITE_PRACK
	m.recv = false
	resp200.recv = true
	st, err = s.sPrack(resp200)
	assert(st == INVITE_PRACK)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sUpdate")

	s.state = INVITE_UPDATE
	m.recv = true
	cancel.recv = true
	st, err = s.sUpdate(cancel)
	assert(st == INVITE_ERR)
	assert(err == nil)
	ct = []check{
		check{typ: CANCEL, code: 200, recv: false},
		check{typ: INVITE, code: 487, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_UPDATE
	update200.recv = true
	st, err = s.sUpdate(update200)
	assert(st == INVITE_18X)
	assert(err == nil)

	s.state = INVITE_UPDATE
	m.recv = true
	resp200.recv = false
	st, err = s.sUpdate(resp200)
	assert(st == INVITE_UPDATE)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++s200")

	s.state = INVITE_200
	s.inviteRecv = true
	ack.recv = true
	st, err = s.s200(ack)
	assert(st == INVITE_ACK)
	assert(err == nil)

	s.state = INVITE_200
	bye.recv = true
	st, err = s.s200(bye)
	assert(st == INVITE_END)
	assert(err == nil)
	ct = []check{
		check{typ: BYE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_200
	reinvite.recv = false
	st, err = s.s200(reinvite)
	assert(st == INVITE_REINV)
	assert(err == nil)
	assert(!s.inviteRecv)

	s.state = INVITE_200
	m.recv = true
	update.recv = true
	st, err = s.s200(update)
	assert(st == INVITE_200)
	assert(err.Error() == "Ignore")
	ct = []check{
		check{typ: UPDATE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_200
	m.recv = false
	update200.recv = true
	st, err = s.s200(update200)
	assert(st == INVITE_200)
	assert(err.Error() == "Ignore")

	s.state = INVITE_200
	st, err = s.s200(reinvite200)
	assert(st == INVITE_200)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sAck")

	s.state = INVITE_ACK
	bye.recv = false
	st, err = s.sAck(bye)
	assert(st == INVITE_END)
	assert(err == nil)

	s.state = INVITE_ACK
	reinvite.recv = true
	st, err = s.sAck(reinvite)
	assert(st == INVITE_REINV)
	assert(err == nil)
	assert(s.inviteRecv)

	s.state = INVITE_ACK
	m.recv = true
	update.recv = true
	st, err = s.sAck(update)
	assert(st == INVITE_ACK)
	assert(err.Error() == "Ignore")
	ct = []check{
		check{typ: UPDATE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_ACK
	m.recv = false
	update200.recv = true
	st, err = s.sAck(update200)
	assert(st == INVITE_ACK)
	assert(err.Error() == "Ignore")

	s.state = INVITE_ACK
	st, err = s.sAck(reinvite200)
	assert(st == INVITE_ACK)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sReInv")

	s.state = INVITE_REINV
	bye.recv = true
	st, err = s.sReInv(bye)
	assert(st == INVITE_END)
	assert(err == nil)
	ct = []check{
		check{typ: BYE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_REINV
	s.inviteRecv = false
	reinvite200.recv = true
	st, err = s.sReInv(reinvite200)
	assert(st == INVITE_RE200)
	assert(err == nil)

	s.state = INVITE_REINV
	m.recv = true
	update.recv = true
	st, err = s.sReInv(update)
	assert(st == INVITE_REINV)
	assert(err.Error() == "Ignore")
	ct = []check{
		check{typ: UPDATE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_REINV
	m.recv = false
	update200.recv = true
	st, err = s.sReInv(update200)
	assert(st == INVITE_REINV)
	assert(err.Error() == "Ignore")

	s.state = INVITE_REINV
	st, err = s.sReInv(reinvite)
	assert(st == INVITE_REINV)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sRe200")

	s.state = INVITE_RE200
	bye.recv = false
	st, err = s.sRe200(bye)
	assert(st == INVITE_END)
	assert(err == nil)

	s.state = INVITE_RE200
	s.inviteRecv = false
	ack.recv = false
	st, err = s.sRe200(ack)
	assert(st == INVITE_ACK)
	assert(err == nil)

	s.state = INVITE_RE200
	m.recv = true
	update.recv = true
	st, err = s.sRe200(update)
	assert(st == INVITE_RE200)
	assert(err.Error() == "Ignore")
	ct = []check{
		check{typ: UPDATE, code: 200, recv: false},
	}
	testSessionMsg(ct, init)

	s.state = INVITE_RE200
	m.recv = false
	update200.recv = true
	st, err = s.sRe200(update200)
	assert(st == INVITE_RE200)
	assert(err.Error() == "Ignore")

	s.state = INVITE_RE200
	st, err = s.sRe200(reinvite)
	assert(st == INVITE_RE200)
	assert(err.Error() == "Unexpected msg")

	fmt.Println("++++++++++sErr")

	s.state = INVITE_ERR
	m.recv = true
	ack.recv = true
	st, err = s.sErr(ack)
	assert(st == INVITE_END)
	assert(err.Error() == "Ignore")
}

// INVITE flow Test

func testSession(m *JSIP, ct []check, to time.Duration) {
	init := &jsipSessionInit{
		sessionFailureCount: 3,
		sessionTimer:        time.Second * 9,
		prTimer:             time.Second * 3,
		transTimer:          time.Second * 1,
		qsize:               1024,
		msg:                 make(chan *JSIP, 1024),
		term:                make(chan string, 1024),
	}

	ss := createSession(m, init, log)
	msg := <-init.msg
	if msg.recv {
		fmt.Println("	recv msg:", msg.String())
	} else {
		fmt.Println("	send msg:", msg.String())
	}
	assert(m == msg)

	start := time.Now()

	for _, c := range ct {
		if c.timeout != 0 {
			time.Sleep(time.Duration(c.timeout) * time.Second)
		}

		if c.msg != nil {
			ss.onMsg(c.msg)
		}

		if c.ignore {
			continue
		}

		msg := <-init.msg
		if msg.recv {
			fmt.Println("	recv msg:", msg.String())
		} else {
			fmt.Println("	send msg:", msg.String())
		}
		assert(msg.Type == c.typ)
		assert(msg.Code == c.code)
		assert(msg.recv == c.recv)
	}

	select {
	case msg := <-init.msg:
		fmt.Println("	recv unexpected msg:", msg.String())
		assert(false)
	case tid := <-init.term:
		if to != time.Duration(0) {
			d := time.Since(start)
			assert(int(d.Seconds()) == int(to.Seconds()))
		}
		fmt.Println("	Session Term", tid)
	}
	return

}

func TestErrResp(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestErrResp")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")

	resp100 := JSIPMsgRes(m, 100)
	resp180 := JSIPMsgRes(m, 180)
	resp408 := JSIPMsgRes(m, 408)

	ack := JSIPMsgAck(resp408)

	fmt.Println("++++++++++Recv")
	m.recv = true
	resp100.recv = false
	resp180.recv = false
	resp408.recv = false
	ack.recv = true

	ct := []check{
		check{msg: resp100, ignore: true},
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{msg: resp408, typ: INVITE, code: 408, recv: false},
		check{msg: ack, ignore: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)

	fmt.Println("++++++++++Send")
	m.recv = false
	resp100.recv = true
	resp180.recv = true
	resp408.recv = true

	ct = []check{
		check{msg: resp100, ignore: true},
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: resp408, typ: ACK, code: 0, recv: false},
		check{typ: INVITE, code: 408, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)
}

func TestTimeoutBefore200(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestTimeoutBefore200")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	resp180 := JSIPMsgRes(m, 408)
	ack := JSIPMsgAck(resp180)
	resp180.Code = 180

	fmt.Println("++++++++++Recv")
	m.recv = true
	resp180.recv = false
	ack.recv = true

	ct := []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{typ: CANCEL, code: 0, recv: true},
		check{typ: INVITE, code: 408, recv: false},
		check{msg: ack, ignore: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 3*time.Second)

	fmt.Println("++++++++++Recv ACK timeout")
	m.recv = true
	resp180.recv = false
	ack.recv = true

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{typ: CANCEL, code: 0, recv: true},
		check{typ: INVITE, code: 408, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 4*time.Second)

	fmt.Println("++++++++++Send")
	m.recv = false
	resp180.recv = true
	ack.recv = false

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{typ: INVITE, code: 408, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 3*time.Second)
}

func TestCancel(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestCancel")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	cancel := JSIPMsgCancel(m)
	resp487 := JSIPMsgRes(m, 487)
	ack := JSIPMsgAck(resp487)

	fmt.Println("++++++++++Recv")
	m.recv = true
	cancel.recv = true
	ack.recv = true

	ct := []check{
		check{msg: cancel, ignore: true},
		check{typ: CANCEL, code: 200, recv: false},
		check{typ: INVITE, code: 487, recv: false},
		check{typ: CANCEL, code: 0, recv: true},
		check{msg: ack, ignore: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)

	fmt.Println("++++++++++Send")
	m.recv = false
	cancel.recv = false
	resp487.recv = true

	ct = []check{
		check{msg: cancel, typ: CANCEL, code: 0, recv: false},
		check{msg: resp487, ignore: true},
		check{typ: ACK, code: 0, recv: false},
		check{typ: INVITE, code: 487, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)
}

func TestINVITE(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestINVITE")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")

	resp180 := JSIPMsgRes(m, 180)

	prack := JSIPMsgReq(PRACK, "jsip.com", "jsip", "jsip", "123456")
	pr200 := JSIPMsgRes(prack, 200)

	update := JSIPMsgUpdate(m)
	up200 := JSIPMsgRes(update, 200)

	resp200 := JSIPMsgRes(m, 200)

	ack := JSIPMsgAck(resp200)

	reinvite := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	re200 := JSIPMsgRes(reinvite, 200)
	reack := JSIPMsgAck(re200)

	bye := JSIPMsgBye(m)

	fmt.Println("++++++++++Recv")
	m.recv = true
	resp180.recv = false
	prack.recv = true
	pr200.recv = false
	update.recv = false
	up200.recv = true
	resp200.recv = false
	ack.recv = true
	reinvite.recv = false
	re200.recv = true
	reack.recv = false
	bye.recv = true

	ct := []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{msg: prack, typ: PRACK, code: 0, recv: true},
		check{msg: pr200, typ: PRACK, code: 200, recv: false},
		check{msg: update, typ: UPDATE, code: 0, recv: false},
		check{msg: up200, typ: UPDATE, code: 200, recv: true},
		check{msg: resp200, typ: INVITE, code: 200, recv: false},
		check{msg: ack, typ: ACK, code: 0, recv: true},
		check{msg: reinvite, typ: INVITE, code: 0, recv: false},
		check{msg: re200, typ: INVITE, code: 200, recv: true},
		check{msg: reack, typ: ACK, code: 0, recv: false},
		check{msg: bye, ignore: true},
		check{typ: BYE, code: 200, recv: false},
		check{typ: BYE, code: 0, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)

	fmt.Println("++++++++++Recv no ACK")
	m.recv = true
	resp180.recv = false
	prack.recv = true
	pr200.recv = false
	update.recv = true
	up200.recv = false
	resp200.recv = false
	reinvite.recv = true
	re200.recv = false
	reack.recv = true
	bye.recv = false

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{msg: prack, typ: PRACK, code: 0, recv: true},
		check{msg: pr200, typ: PRACK, code: 200, recv: false},
		check{msg: update, typ: UPDATE, code: 0, recv: true},
		check{msg: up200, typ: UPDATE, code: 200, recv: false},
		check{msg: resp200, typ: INVITE, code: 200, recv: false},
		check{msg: reinvite, typ: INVITE, code: 0, recv: true},
		check{msg: re200, typ: INVITE, code: 200, recv: false},
		check{msg: reack, typ: ACK, code: 0, recv: true},
		check{msg: bye, typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)

	fmt.Println("++++++++++Send")
	m.recv = false
	resp180.recv = true
	prack.recv = false
	pr200.recv = true
	update.recv = true
	up200.recv = false
	resp200.recv = true
	ack.recv = false
	reinvite.recv = true
	re200.recv = false
	reack.recv = true
	bye.recv = true

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: prack, typ: PRACK, code: 0, recv: false},
		check{msg: pr200, typ: PRACK, code: 200, recv: true},
		check{msg: update, typ: UPDATE, code: 0, recv: true},
		check{msg: up200, typ: UPDATE, code: 200, recv: false},
		check{msg: resp200, typ: INVITE, code: 200, recv: true},
		check{msg: ack, typ: ACK, code: 0, recv: false},
		check{msg: reinvite, typ: INVITE, code: 0, recv: true},
		check{msg: re200, typ: INVITE, code: 200, recv: false},
		check{msg: reack, typ: ACK, code: 0, recv: true},
		check{msg: bye, ignore: true},
		check{typ: BYE, code: 200, recv: false},
		check{typ: BYE, code: 0, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)

	fmt.Println("++++++++++Send no ACK")
	m.recv = false
	resp180.recv = true
	prack.recv = false
	pr200.recv = true
	update.recv = false
	up200.recv = true
	resp200.recv = true
	reinvite.recv = false
	re200.recv = true
	reack.recv = false
	bye.recv = false

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: prack, typ: PRACK, code: 0, recv: false},
		check{msg: pr200, typ: PRACK, code: 200, recv: true},
		check{msg: update, typ: UPDATE, code: 0, recv: false},
		check{msg: up200, typ: UPDATE, code: 200, recv: true},
		check{msg: resp200, typ: INVITE, code: 200, recv: true},
		check{msg: reinvite, typ: INVITE, code: 0, recv: false},
		check{msg: re200, typ: INVITE, code: 200, recv: true},
		check{msg: reack, typ: ACK, code: 0, recv: false},
		check{msg: bye, typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 0)
}

func TestSessionTimeout(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestSessionTimeout")

	m := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")

	resp180 := JSIPMsgRes(m, 180)
	resp200 := JSIPMsgRes(m, 200)
	ack := JSIPMsgAck(resp200)

	update := JSIPMsgUpdate(m)
	up200 := JSIPMsgRes(update, 200)

	bye := JSIPMsgBye(m)

	fmt.Println("++++++++++Recv")
	m.recv = true
	resp180.recv = false
	resp200.recv = false
	ack.recv = true
	update.recv = true
	bye.recv = false

	ct := []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{msg: resp200, typ: INVITE, code: 200, recv: false},
		check{msg: ack, typ: ACK, code: 0, recv: true},
		check{msg: update, typ: UPDATE, code: 200, recv: false, timeout: 3},
		check{msg: bye, typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 3*time.Second)

	fmt.Println("++++++++++Recv Timeout")
	m.recv = true
	resp180.recv = false
	resp200.recv = false
	ack.recv = true

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{msg: resp200, typ: INVITE, code: 200, recv: false},
		check{msg: ack, typ: ACK, code: 0, recv: true},
		check{typ: BYE, code: 0, recv: true},
		check{typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 9*time.Second)

	fmt.Println("++++++++++Send")
	m.recv = false
	resp180.recv = true
	resp200.recv = true
	ack.recv = false
	up200.recv = true
	bye.recv = false

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: resp200, typ: INVITE, code: 200, recv: true},
		check{msg: ack, typ: ACK, code: 0, recv: false},
		check{typ: UPDATE, code: 0, recv: false},
		check{msg: up200, ignore: true},
		check{typ: UPDATE, code: 0, recv: false},
		check{msg: up200, ignore: true},
		check{msg: bye, typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 6*time.Second)

	fmt.Println("++++++++++Send Timeout")
	m.recv = false
	resp180.recv = true
	resp200.recv = true
	ack.recv = false

	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: resp200, typ: INVITE, code: 200, recv: true},
		check{msg: ack, typ: ACK, code: 0, recv: false},
		check{typ: UPDATE, code: 0, recv: false},
		check{typ: UPDATE, code: 0, recv: false},
		check{typ: UPDATE, code: 0, recv: false},
		check{msg: bye, typ: BYE, code: 0, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testSession(m, ct, 9*time.Second)
}
