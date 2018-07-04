package main

import (
	"fmt"
	"time"

	"rtclib"
)

type Slpdemo struct {
	task   *rtclib.Task
	invite *rtclib.JSIP
}

type slpdemoctx struct {
	a int
	b int
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	return &Slpdemo{task: task}
}

func test(jsip interface{}) {
	fmt.Println("!!!!!!!test", time.Now(), jsip.(*rtclib.JSIP))
}

func (slp *Slpdemo) NewSLPCtx() interface{} {
	fmt.Println("!!!!!! NewSLPCtx")
	return &slpdemoctx{a: 10, b: 20}
}

func (slp *Slpdemo) OnLoad(jsip *rtclib.JSIP) {
	fmt.Println("!!!!!! initProcess")
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) {
	fmt.Printf("slp: %p %s %s %s\n",
		slp, jsip.Name(), jsip.DialogueID, time.Now())

	//switch jsip.Type {
	//case rtclib.INVITE:
	//timer := rtclib.NewTimer(5*time.Second, test, jsip)
	//fmt.Println("!!!!!!!!", time.Now(), timer)
	//timer.Reset(10 * time.Second)
	//timer.Stop()
	if jsip.Type == rtclib.INVITE && jsip.Code == 0 {
		slp.invite = jsip

		resp := rtclib.JSIPMsgRes(jsip, 200)
		rtclib.SendMsg(resp)
		slp.task.SetFinished()

		//prack := rtclib.JSIPMsgAck(resp)
		//prack.Type = rtclib.PRACK
		//rtclib.SendMsg(prack)

		//dlg := slp.task.NewDialogueID()
		//invite := rtclib.JSIPMsgClone(jsip, dlg)
		//invite.Router = invite.Router[1:]
		//rtclib.SendMsg(invite)

		//t2 := rtclib.JSIPMsgCancel(invite)
		//rtclib.SendMsg(t2)

		//ack := &rtclib.JSIP{
		//	Type:       rtclib.ACK,
		//	RequestURI: invite.RequestURI,
		//	From:       invite.From,
		//	To:         invite.To,
		//	DialogueID: dlg,
		//	RawMsg:     make(map[string]interface{}),
		//}

		//ack.RawMsg["RelatedID"] = json.Number(strconv.Itoa(1))
		//rtclib.SendMsg(ack)

		//fmt.Println("+++++++++", resp.Transaction)
		//cancel := rtclib.JSIPMsgCancel(resp.Transaction)
		//rtclib.SendMsg(cancel)
	}

	if (jsip.Type == rtclib.PRACK || jsip.Type == rtclib.UPDATE) &&
		jsip.Code == 0 {

		resp := rtclib.JSIPMsgRes(jsip, 200)
		rtclib.SendMsg(resp)

		if jsip.Type == rtclib.UPDATE {
			update := rtclib.JSIPMsgClone(jsip, jsip.DialogueID)
			rtclib.SendMsg(update)
		}
	}

	if jsip.Type == rtclib.INVITE && jsip.Code >= 100 && jsip.Code < 200 {
		prack := rtclib.JSIPMsgAck(jsip)
		prack.Type = rtclib.PRACK
		rtclib.SendMsg(prack)

		cancel := rtclib.JSIPMsgCancel(slp.invite)
		cancel.DialogueID = jsip.DialogueID
		rtclib.SendMsg(cancel)
	}

	if jsip.Type == rtclib.INVITE && jsip.Code == 200 {
		ack := rtclib.JSIPMsgAck(jsip)
		rtclib.SendMsg(ack)
	}

	return
}
