// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Transaction Test Case

package rtclib

import (
	"fmt"
	"testing"
	"time"

	"github.com/alexwoo/golib"
)

var log = golib.NewLog("transaction.log")

func TestTransactionCommon(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestTransactionCommon")

	init := &jsipTransInit{
		transTimer: time.Second * 5,
		prTimer:    time.Second * 60,
		qsize:      1024,
		msg:        make(chan *JSIP, 1024),
		term:       make(chan string, 1024),
	}

	if transactionID("test", 1234) != "test:1234" {
		t.Error("transactionID failed")
	}

	invite := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	res := JSIPMsgRes(invite, 200)

	assert(createTransaction(invite, init, log) != nil)
	assert(createTransaction(res, init, log) == nil)
}

func TestTransProcess(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestTransProcess")

	init := &jsipTransInit{
		transTimer: time.Second * 5,
		prTimer:    time.Second * 60,
		qsize:      1024,
		msg:        make(chan *JSIP, 1024),
		term:       make(chan string, 1024),
	}

	invite := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	invite180 := JSIPMsgRes(invite, 180)
	invite180.recv = true
	invite200 := JSIPMsgRes(invite, 200)
	invite200.recv = true
	invite500 := JSIPMsgRes(invite, 500)
	invite500.recv = true

	cancel := JSIPMsgReq(CANCEL, "jsip.com", "jsip", "jsip", "123456")
	resp := JSIPMsgRes(cancel, 200)

	tt := createTransaction(invite, init, log)

	// Error Check
	state, err := tt.transProcess(cancel)
	assert(state == TRANS_INIT)
	assert(err.Error() == "Unexpected msg")

	state, err = tt.transProcess(resp)
	assert(state == TRANS_INIT)
	assert(err.Error() == "Unexpected response")

	resp.Type = INVITE
	resp.CSeq = 1
	state, err = tt.transProcess(resp)
	assert(state == TRANS_INIT)
	assert(err.Error() == "Unexpected CSeq")

	resp.CSeq = invite.CSeq
	state, err = tt.transProcess(resp)
	assert(state == TRANS_INIT)
	assert(err.Error() == "Unexpected direction")

	// TRANS_INIT
	tt.state = TRANS_INIT
	state, err = tt.transProcess(invite180)
	assert(state == TRANS_PROVISIONALRESP)
	assert(err == nil)

	tt.state = TRANS_INIT
	state, err = tt.transProcess(invite200)
	assert(state == TRANS_SUCCESSESP)
	assert(err == nil)

	tt.state = TRANS_INIT
	state, err = tt.transProcess(invite500)
	assert(state == TRANS_ERRRESP)
	assert(err == nil)

	// TRANS_PROVISIONALRESP
	tt.state = TRANS_PROVISIONALRESP
	state, err = tt.transProcess(invite180)
	assert(state == TRANS_PROVISIONALRESP)
	assert(err == nil)

	tt.state = TRANS_PROVISIONALRESP
	state, err = tt.transProcess(invite200)
	assert(state == TRANS_SUCCESSESP)
	assert(err == nil)

	tt.state = TRANS_PROVISIONALRESP
	state, err = tt.transProcess(invite500)
	assert(state == TRANS_ERRRESP)
	assert(err == nil)

	// TRANS_SUCCESSESP
	tt.state = TRANS_SUCCESSESP
	state, err = tt.transProcess(invite180)
	assert(state == TRANS_SUCCESSESP)
	assert(err.Error() == "Unexpected provisional response")

	tt.state = TRANS_SUCCESSESP
	state, err = tt.transProcess(invite200)
	assert(state == TRANS_SUCCESSESP)
	assert(err.Error() == "Unexpected final response")

	tt.state = TRANS_SUCCESSESP
	state, err = tt.transProcess(invite500)
	assert(state == TRANS_SUCCESSESP)
	assert(err.Error() == "Unexpected final response")

	// TRANS_ERRRESP
	tt.state = TRANS_ERRRESP
	state, err = tt.transProcess(invite180)
	assert(state == TRANS_ERRRESP)
	assert(err.Error() == "Unexpected provisional response")

	tt.state = TRANS_ERRRESP
	state, err = tt.transProcess(invite200)
	assert(state == TRANS_ERRRESP)
	assert(err.Error() == "Unexpected final response")

	tt.state = TRANS_ERRRESP
	state, err = tt.transProcess(invite500)
	assert(state == TRANS_ERRRESP)
	assert(err.Error() == "Unexpected final response")
}

func testTransaction(m *JSIP, ct []check, to time.Duration) {
	init := &jsipTransInit{
		transTimer: time.Second * 1,
		prTimer:    time.Second * 3,
		qsize:      1024,
		msg:        make(chan *JSIP, 1024),
		term:       make(chan string, 1024),
	}

	tt := createTransaction(m, init, log)
	msg := <-init.msg
	if msg.recv {
		fmt.Println("	recv msg:", msg.String())
	} else {
		fmt.Println("	send msg:", msg.String())
	}
	assert(m == msg)

	start := time.Now()

	for _, c := range ct {
		if c.msg != nil {
			tt.onMsg(c.msg)
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
		fmt.Println("	Transaction Term", tid)
	}
	return

}

func TestINVITETransaction(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestINVITETransaction")

	msg := JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456")
	resp := JSIPMsgRes(msg, 200)

	resp180 := JSIPMsgRes(msg, 180)
	resp200 := JSIPMsgRes(msg, 200)
	resp500 := JSIPMsgRes(msg, 500)

	fmt.Println("++++++++++Test trans process error")
	ct := []check{
		check{msg: resp, ignore: true},
		check{typ: INVITE, code: 408, recv: true},
	}
	testTransaction(msg, ct, 6*time.Second)

	fmt.Println("++++++++++Send")
	msg.recv = false
	resp180.recv = true
	resp200.recv = true
	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{msg: resp200, typ: INVITE, code: 200, recv: true},
	}
	testTransaction(msg, ct, 0)

	fmt.Println("++++++++++Send Error")
	msg.recv = false
	resp500.recv = true
	ct = []check{
		check{msg: resp500, typ: INVITE, code: 500, recv: true},
	}
	testTransaction(msg, ct, 0)

	fmt.Println("++++++++++Send Timeout 1")
	msg.recv = false
	ct = []check{
		check{typ: INVITE, code: 408, recv: true},
	}
	testTransaction(msg, ct, 6*time.Second)

	fmt.Println("++++++++++Send Timeout 2")
	msg.recv = false
	resp180.recv = true
	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: true},
		check{typ: INVITE, code: 408, recv: true},
	}
	testTransaction(msg, ct, 6*time.Second)

	fmt.Println("++++++++++Recv")
	msg.recv = true
	resp200.recv = false
	ct = []check{
		check{msg: resp200, typ: INVITE, code: 200, recv: false},
	}
	testTransaction(msg, ct, 0)

	fmt.Println("++++++++++Recv Error")
	msg.recv = true
	resp500.recv = false
	ct = []check{
		check{msg: resp500, typ: INVITE, code: 500, recv: false},
	}
	testTransaction(msg, ct, 0)

	fmt.Println("++++++++++Recv Timeout 1")
	msg.recv = true
	ct = []check{
		check{typ: INVITE, code: 408, recv: false},
	}
	testTransaction(msg, ct, 6*time.Second)

	fmt.Println("++++++++++Recv Timeout 2")
	msg.recv = true
	resp180.recv = false
	ct = []check{
		check{msg: resp180, typ: INVITE, code: 180, recv: false},
		check{typ: INVITE, code: 408, recv: false},
	}
	testTransaction(msg, ct, 6*time.Second)
}

func TestREGISTERTransaction(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestREGISTERTransaction")

	msg := JSIPMsgReq(REGISTER, "jsip.com", "jsip", "jsip", "123456")
	resp := JSIPMsgRes(msg, 200)

	resp180 := JSIPMsgRes(msg, 180)
	resp200 := JSIPMsgRes(msg, 200)
	resp500 := JSIPMsgRes(msg, 500)

	fmt.Println("++++++++++Test trans process error")
	ct := []check{
		check{msg: resp, ignore: true},
		check{typ: REGISTER, code: 408, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send")
	msg.recv = false
	resp180.recv = true
	resp200.recv = true
	ct = []check{
		check{msg: resp180, typ: REGISTER, code: 180, recv: true},
		check{msg: resp200, typ: REGISTER, code: 200, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Error")
	msg.recv = false
	resp500.recv = true
	ct = []check{
		check{msg: resp500, typ: REGISTER, code: 500, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Timeout 1")
	msg.recv = false
	ct = []check{
		check{typ: REGISTER, code: 408, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send Timeout 2")
	msg.recv = false
	resp180.recv = true
	ct = []check{
		check{msg: resp180, typ: REGISTER, code: 180, recv: true},
		check{typ: REGISTER, code: 408, recv: true},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 3*time.Second)

	fmt.Println("++++++++++Recv")
	msg.recv = true
	resp200.recv = false
	ct = []check{
		check{msg: resp200, typ: REGISTER, code: 200, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Error")
	msg.recv = true
	resp500.recv = false
	ct = []check{
		check{msg: resp500, typ: REGISTER, code: 500, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Timeout 1")
	msg.recv = true
	ct = []check{
		check{typ: REGISTER, code: 408, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Recv Timeout 2")
	msg.recv = true
	resp180.recv = false
	ct = []check{
		check{msg: resp180, typ: REGISTER, code: 180, recv: false},
		check{typ: REGISTER, code: 408, recv: false},
		check{typ: TERM, code: 0, recv: true},
	}
	testTransaction(msg, ct, 3*time.Second)
}

func TestACKTransaction(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestACKTransaction")

	ack := JSIPMsgReq(ACK, "jsip.com", "jsip", "jsip", "123456")

	fmt.Println("++++++++++Recv")
	ack.recv = false
	testTransaction(ack, []check{}, time.Duration(0))

	fmt.Println("++++++++++Send")
	ack.recv = true
	testTransaction(ack, []check{}, time.Duration(0))
}

func TestCANCELTransaction(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestCANCELTransaction")

	msg := JSIPMsgReq(CANCEL, "jsip.com", "jsip", "jsip", "123456")
	resp := JSIPMsgRes(msg, 200)

	resp180 := JSIPMsgRes(msg, 180)
	resp200 := JSIPMsgRes(msg, 200)
	resp500 := JSIPMsgRes(msg, 500)

	fmt.Println("++++++++++Test trans process error")
	ct := []check{
		check{msg: resp, ignore: true},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send")
	msg.recv = false
	resp180.recv = true
	resp200.recv = true
	ct = []check{
		check{msg: resp180, ignore: true},
		check{msg: resp200, ignore: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Error")
	msg.recv = false
	resp500.recv = true
	ct = []check{
		check{msg: resp500, ignore: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Timeout 1")
	msg.recv = false
	ct = []check{}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send Timeout 2")
	msg.recv = false
	resp180.recv = true
	ct = []check{
		check{msg: resp180, ignore: true},
	}
	testTransaction(msg, ct, 3*time.Second)

	fmt.Println("++++++++++Recv")
	msg.recv = true
	resp200.recv = false
	ct = []check{
		check{msg: resp200, typ: CANCEL, code: 200, recv: false},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Error")
	msg.recv = true
	resp500.recv = false
	ct = []check{
		check{msg: resp500, typ: CANCEL, code: 500, recv: false},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Timeout 1")
	msg.recv = true
	ct = []check{
		check{typ: CANCEL, code: 408, recv: false},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Recv Timeout 2")
	msg.recv = true
	resp180.recv = false
	ct = []check{
		check{msg: resp180, typ: CANCEL, code: 180, recv: false},
		check{typ: CANCEL, code: 408, recv: false},
	}
	testTransaction(msg, ct, 3*time.Second)
}

func TestBYETransaction(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestBYETransaction")

	msg := JSIPMsgReq(BYE, "jsip.com", "jsip", "jsip", "123456")
	resp := JSIPMsgRes(msg, 200)

	resp180 := JSIPMsgRes(msg, 180)
	resp200 := JSIPMsgRes(msg, 200)
	resp500 := JSIPMsgRes(msg, 500)

	fmt.Println("++++++++++Test trans process error")
	ct := []check{
		check{msg: resp, ignore: true},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send")
	msg.recv = false
	resp180.recv = true
	resp200.recv = true
	ct = []check{
		check{msg: resp180, ignore: true},
		check{msg: resp200, ignore: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Error")
	msg.recv = false
	resp500.recv = true
	ct = []check{
		check{msg: resp500, ignore: true},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Send Timeout 1")
	msg.recv = false
	ct = []check{}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Send Timeout 2")
	msg.recv = false
	resp180.recv = true
	ct = []check{
		check{msg: resp180, ignore: true},
	}
	testTransaction(msg, ct, 3*time.Second)

	fmt.Println("++++++++++Recv")
	msg.recv = true
	resp200.recv = false
	ct = []check{
		check{msg: resp200, typ: BYE, code: 200, recv: false},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Error")
	msg.recv = true
	resp500.recv = false
	ct = []check{
		check{msg: resp500, typ: BYE, code: 500, recv: false},
	}
	testTransaction(msg, ct, 0*time.Second)

	fmt.Println("++++++++++Recv Timeout 1")
	msg.recv = true
	ct = []check{
		check{typ: BYE, code: 408, recv: false},
	}
	testTransaction(msg, ct, 1*time.Second)

	fmt.Println("++++++++++Recv Timeout 2")
	msg.recv = true
	resp180.recv = false
	ct = []check{
		check{msg: resp180, typ: BYE, code: 180, recv: false},
		check{typ: BYE, code: 408, recv: false},
	}
	testTransaction(msg, ct, 3*time.Second)
}
