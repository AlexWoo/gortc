// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Message Test Case

package rtclib

import (
	"fmt"
	"testing"
)

func TestMarshal(t *testing.T) {
	fmt.Println("!!!!!!!!!TestMarshal")

	if _, err := (&JSIP{
		Type:   JSIPType(100),
		rawMsg: make(map[string]interface{}),
	}).Marshal(); err.Error() != "Unknow Type" {
		t.Error(err.Error(), "test error")
	}

	if _, err := (&JSIP{
		Type:   INVITE,
		rawMsg: make(map[string]interface{}),
	}).Marshal(); err.Error() != "No RequestURI" {
		t.Error(err.Error(), "test error")
	}

	if _, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err.Error() != "No From" {
		t.Error(err.Error(), "test error")
	}

	if _, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err.Error() != "No To" {
		t.Error(err.Error(), "test error")
	}

	if _, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err.Error() != "No CSeq" {
		t.Error(err.Error(), "test error")
	}

	if _, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err.Error() != "No DialogueID" {
		t.Error(err.Error(), "test error")
	}

	if b, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err != nil {
		t.Error(err.Error(), "unexpected test error")
	} else {
		fmt.Println(string(b))
	}

	if b, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg: map[string]interface{}{
			"test":      "abcd",
			"RelatedID": 111,
			"Router":    "eeee", // Router will be deleted
		},
	}).Marshal(); err != nil {
		t.Error(err.Error(), "unexpected test error")
	} else {
		fmt.Println(string(b))
	}

	if b, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		Router:     []string{"router1"}, // add one Router
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err != nil {
		t.Error(err.Error(), "unexpected test error")
	} else {
		fmt.Println(string(b))
	}

	if b, err := (&JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		Router:     []string{"router1", "router2"}, // add two Router
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err != nil {
		t.Error(err.Error(), "unexpected test error")
	} else {
		fmt.Println(string(b))
	}

	if _, err := (&JSIP{
		Type:       INVITE,
		Code:       2,
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err.Error() != "Invalid Code" {
		t.Error(err.Error(), "test error")
	}

	if b, err := (&JSIP{
		Type:       INVITE,
		Code:       199,
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}).Marshal(); err != nil {
		t.Error(err.Error(), "unexpected test error")
	} else {
		fmt.Println(string(b))
	}
}

func TestUnmarshal(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestUnmarshal")

	t1 := `{"CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","To":"jsip","Type":"INVITE","Router":"router1"}`
	req1 := &JSIP{}
	if err := req1.Unmarshal([]byte(t1)); err != nil {
		t.Error("Unmarshal JSIP Request Failure")
	}
	assert(req1.Type == INVITE)
	assert(req1.RequestURI == "jsip.com")
	assert(req1.From == "jsip")
	assert(req1.To == "jsip")
	assert(req1.CSeq == 1)
	assert(req1.DialogueID == "test1")
	assert(req1.Router[0] == "router1")

	t2 := `{"CSeq":1,"Code":199,"Desc":"User Defined Provisional Response","DialogueID":"test1","From":"jsip","To":"jsip","Type":"RESPONSE"}`
	res2 := &JSIP{}
	if err := res2.Unmarshal([]byte(t2)); err != nil {
		t.Error("Unmarshal JSIP Response Failure")
	}
	assert(res2.Code == 199)
	assert(res2.From == "jsip")
	assert(res2.To == "jsip")
	assert(res2.CSeq == 1)
	assert(res2.DialogueID == "test1")

	t3 := `{"Body":["aaa", "bbb", "ccc": {"ddd": "eee", "fff": 100}],"CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","To":"jsip","Type":"INVITE","Router":"router1, router2"}`
	req3 := &JSIP{}
	if err := req3.Unmarshal([]byte(t3)); err != nil {
		t.Error("Unmarshal JSIP Request Failure")
	}
	assert(req3.DialogueID == "test1")
	assert(req3.Router[0] == "router1")
	assert(req3.Router[1] == "router2")
	fmt.Println(req3.Body)

	t4 := `["aaa", "bbb", "ddd"]`
	req4 := &JSIP{}
	if err := req4.Unmarshal([]byte(t4)); err.Error() != "raw is not json object" {
		t.Error(err.Error(), "test error")
	}

	t5 := `{"CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","To":"jsip","Router":"router1"}`
	req5 := &JSIP{}
	if err := req5.Unmarshal([]byte(t5)); err.Error() != "Type error" {
		t.Error(err.Error(), "test error")
	}

	t6 := `{"Type":"Hello","CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","To":"jsip","Router":"router1"}`
	req6 := &JSIP{}
	if err := req6.Unmarshal([]byte(t6)); err.Error() != "Unknown Type" {
		t.Error(err.Error(), "test error")
	}

	t7 := `{"Type":"INVITE","CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":1,"To":"jsip","Router":"router1"}`
	req7 := &JSIP{}
	if err := req7.Unmarshal([]byte(t7)); err.Error() != "Request-URI error" {
		t.Error(err.Error(), "test error")
	}

	t8 := `{"Type":"RESPONSE","CSeq":1,"DialogueID":"test1","From":"jsip","To":"jsip","Router":"router1"}`
	res8 := &JSIP{}
	if err := res8.Unmarshal([]byte(t8)); err.Error() != "Code error" {
		t.Error(err.Error(), "test error")
	}

	t9 := `{"Type":"RESPONSE","Code":99,"CSeq":1,"DialogueID":"test1","From":"jsip","To":"jsip","Router":"router1"}`
	res9 := &JSIP{}
	if err := res9.Unmarshal([]byte(t9)); err.Error() != "Invalid Code" {
		t.Error(err.Error(), "test error")
	}

	t10 := `{"Type":"INVITE","CSeq":1,"DialogueID":"test1","From":true,"Request-URI":"jsip.com","To":"jsip","Router":"router1"}`
	req10 := &JSIP{}
	if err := req10.Unmarshal([]byte(t10)); err.Error() != "From error" {
		t.Error(err.Error(), "test error")
	}

	t11 := `{"Type":"INVITE","CSeq":1,"DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","Router":"router1"}`
	req11 := &JSIP{}
	if err := req11.Unmarshal([]byte(t11)); err.Error() != "To error" {
		t.Error(err.Error(), "test error")
	}

	t12 := `{"Type":"INVITE","CSeq":"1","DialogueID":"test1","From":"jsip","Request-URI":"jsip.com","To":"jsip","Router":"router1"}`
	req12 := &JSIP{}
	if err := req12.Unmarshal([]byte(t12)); err.Error() != "CSeq error" {
		t.Error(err.Error(), "test error")
	}

	t13 := `{"Type":"INVITE","CSeq":1,"From":"jsip","Request-URI":"jsip.com","To":"jsip","Router":"router1"}`
	req13 := &JSIP{}
	if err := req13.Unmarshal([]byte(t13)); err.Error() != "DialogueID error" {
		t.Error(err.Error(), "test error")
	}
}

func TestJSIPCommon(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPCommon")

	req := &JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}

	res := &JSIP{
		Type:       INVITE,
		Code:       200,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}

	term := JSIPMsgTerm("test1")

	assert(JSIPMsgReq(INVITE, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(ACK, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(BYE, "jsip.co)m", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(CANCEL, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(!JSIPMsgReq(REGISTER, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(!JSIPMsgReq(OPTIONS, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(INFO, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(UPDATE, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(JSIPMsgReq(PRACK, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(!JSIPMsgReq(SUBSCRIBE, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(!JSIPMsgReq(MESSAGE, "jsip.com", "jsip", "jsip", "123456").inviteSession())
	assert(!JSIPMsgReq(NOTIFY, "jsip.com", "jsip", "jsip", "123456").inviteSession())

	if req.Name() != "INVITE" {
		t.Error("Request Name error")
	}

	if res.Name() != "INVITE_200" {
		t.Error("Response Name error")
	}

	if req.String() != "INVITE RequestURI: jsip.com From: jsip To: jsip CSeq: 1 DialogueID: test1" {
		t.Error("Request String error")
	}

	if res.String() != "INVITE_200 From: jsip To: jsip CSeq: 1 DialogueID: test1" {
		t.Error("Response String error")
	}

	if term.String() != "TERM DialogueID: test1" {
		t.Error("TERM String error")
	}

	req.SetInt("RelatedID", 10)
	if i, ok := req.GetInt("RelatedID"); !ok {
		t.Error("GetInt error")
	} else if i != 10 {
		t.Error("GetInt get wrong value")
	}

	req.SetUint("AAA", 11)
	if i, ok := req.GetUint("AAA"); !ok {
		t.Error("GetInt error")
	} else if i != 11 {
		t.Error("GetUint get wrong value")
	}

	req.SetString("Hello", "World")
	if s, ok := req.GetString("Hello"); !ok {
		t.Error("GetInt error")
	} else if s != "World" {
		t.Error("GetString get wrong value")
	}

	req.DelHeader("Hello")
	if _, ok := req.GetString("Hello"); ok {
		t.Error("DelHeader error")
	}

	b, _ := req.Marshal()
	fmt.Println(string(b))
}

func TestJSIPConstruct(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPConstruct")

	// Test Clonse
	req := &JSIP{
		Type:       INVITE,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg: map[string]interface{}{
			"Hello": "World",
			"AAA":   float64(10),
		},
	}

	res := &JSIP{
		Type:       INVITE,
		Code:       200,
		RequestURI: "jsip.com",
		From:       "jsip",
		To:         "jsip",
		CSeq:       1,
		DialogueID: "test1",
		rawMsg:     make(map[string]interface{}),
	}

	term := JSIPMsgTerm("term1")
	assert(term.Type == TERM)
	assert(term.DialogueID == "term1")
	assert(term.recv)

	clone1 := JSIPMsgClone(req, "clone1")
	assert(req.Type == clone1.Type)
	assert(req.RequestURI == clone1.RequestURI)
	assert(req.From == clone1.From)
	assert(req.To == clone1.To)
	assert(req.CSeq != clone1.CSeq)
	assert("clone1" == clone1.DialogueID)
	if h, ok := clone1.GetString("Hello"); !ok {
		t.Error("Get Clone1 Hello failed")
	} else {
		assert(h == "World")
	}
	if h, ok := clone1.GetInt("AAA"); !ok {
		t.Error("Get Clone1 AAA failed")
	} else {
		assert(h == 10)
	}

	clone2 := JSIPMsgClone(res, "clone2")
	assert(res.Type == clone2.Type)
	assert(res.Code == clone2.Code)
	assert(res.From == clone2.From)
	assert(res.To == clone2.To)
	assert(res.CSeq != clone2.CSeq)
	assert("clone2" == clone2.DialogueID)

	// Test Req
	req1 := JSIPMsgReq(INVITE, "test.com", "Alex", "test.com", "test")
	assert(req1.Type == INVITE)
	assert(req1.RequestURI == "test.com")
	assert(req1.From == "Alex")
	assert(req1.To == "test.com")
	assert(req1.DialogueID == "test")

	// Test Res
	res1 := JSIPMsgRes(req1, 200)
	assert(res1.Type == req1.Type)
	assert(res1.Code == 200)
	assert(res1.From == req1.From)
	assert(res1.To == req1.To)
	assert(res1.CSeq == req1.CSeq)
	assert(res1.DialogueID == req1.DialogueID)

	if res2 := JSIPMsgRes(req1, 10); res2 != nil {
		t.Error("Error code exception test failed")
	}

	if res3 := JSIPMsgRes(res1, 200); res3 != nil {
		t.Error("ref msg exception test failed")
	}

	// Test Ack
	ack1 := JSIPMsgAck(res1)
	assert(ack1.Type == ACK)
	assert(ack1.RequestURI == res1.To)
	assert(ack1.From == res1.From)
	assert(ack1.To == res1.To)
	assert(ack1.CSeq != 0)
	assert(ack1.DialogueID == res1.DialogueID)
	if h, ok := ack1.GetUint("RelatedID"); !ok {
		t.Error("ACK has no RelatedID")
	} else if h != res1.CSeq {
		t.Error("ACK RelatedID Error")
	}

	res4 := JSIPMsgRes(req1, 180)
	if ack2 := JSIPMsgAck(res4); ack2 != nil {
		t.Error("ref msg exception test failed")
	}

	res4.Code = 200
	res4.Type = REGISTER
	if ack3 := JSIPMsgAck(res4); ack3 != nil {
		t.Error("ref msg exception test failed")
	}

	// Test Cancel
	cancel1 := JSIPMsgCancel(req1)
	assert(cancel1.Type == CANCEL)
	assert(cancel1.RequestURI == req1.RequestURI)
	assert(cancel1.From == req1.From)
	assert(cancel1.To == req1.To)
	assert(cancel1.DialogueID == req1.DialogueID)
	if h, ok := cancel1.GetUint("RelatedID"); !ok {
		t.Error("CANCEL has no RelatedID")
	} else if h != req1.CSeq {
		t.Error("CACNEL RelatedID Error")
	}

	req1.Type = REGISTER
	cancel2 := JSIPMsgCancel(req1)
	assert(cancel2 == nil)

	// Test BYE
	bye1 := JSIPMsgBye(req)
	assert(bye1.Type == BYE)
	assert(bye1.RequestURI == req.RequestURI)
	assert(bye1.From == req.From)
	assert(bye1.To == req.To)
	assert(bye1.DialogueID == req.DialogueID)

	bye2 := JSIPMsgBye(res)
	assert(bye2 == nil)

	// Test UPDATE
	update1 := JSIPMsgUpdate(req)
	assert(update1.Type == UPDATE)
	assert(update1.RequestURI == req.RequestURI)
	assert(update1.From == req.From)
	assert(update1.To == req.To)
	assert(update1.DialogueID == req.DialogueID)

	update2 := JSIPMsgUpdate(bye1)
	assert(update2 == nil)
}
