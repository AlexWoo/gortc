// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Test

package main

import (
	"rtclib"
	"sync"
	"time"
)

type ctx struct {
}

type TestSLP struct {
	task    *rtclib.Task
	handler func(*rtclib.JSIP)
	count   uint8
	req     *rtclib.JSIP
}

var (
	once sync.Once
	c    *ctx
)

func GetInstance(task *rtclib.Task) rtclib.SLP {
	slp := &TestSLP{
		task: task,
	}

	return slp
}

func (slp *TestSLP) NewSLPCtx() interface{} {
	once.Do(func() {
		c = &ctx{}
	})

	return c
}

func (slp *TestSLP) Process(m *rtclib.JSIP) {
	if slp.handler != nil {
		slp.handler(m)
		return
	}

	name := m.RequestURI
	handler := slp.findHandler(name)
	if handler == nil {
		slp.task.LogError("Cannot find test case, %s", name)
		rtclib.SendMsg(rtclib.JSIPMsgRes(m, 404))

		slp.task.SetFinished()

		return
	}

	slp.handler = handler
	slp.task.LogInfo("Start test case %s", name)

	handler(m)
}

func (slp *TestSLP) OnLoad(m *rtclib.JSIP) {
	slp.task.LogInfo("Onload")
}

func assert(b bool) {
	if !b {
		panic("assert failed")
	}
}

func (slp *TestSLP) findHandler(name string) func(*rtclib.JSIP) {
	switch name {
	case "Recv_MESSAGE":
		return slp.RecvMESSAGE

	case "Recv_INVITE":
		return slp.RecvINVITE

	case "Recv_INVITE_Timeout":
		return slp.RecvINVITETimeout

	case "Recv_MESSAGE_Timeout":
		return slp.RecvMESSAGETimeout

	case "Recv_INVITE_Session_Timeout":
		return slp.RecvINVITESessionTimeout

	case "Recv_CANCEL":
		return slp.RecvCANCEL

	case "Send_Error_Msg":
		return slp.SendErrorMsg

	case "Send_BYE_481":
		return slp.SendBYE481

	case "Send_ACK_No_Session":
		return slp.SendACKNoSession

	case "Send_CANCEL_No_Session":
		return slp.SendCANCELNoSession

	case "Send_Resp_No_Session":
		return slp.SendRespNoSession

	case "Send_UPDATE_No_Session":
		return slp.SendUPDATENoSession

	case "Send_MESSAGE":
		return slp.SendMessage

	case "Send_INVITE":
		return slp.SendINVITE

	case "Send_INVITE_Timeout":
		return slp.SendINVITETimeout

	case "Send_MESSAGE_Timeout":
		return slp.SendMESSAGETimeout

	case "Send_INVITE_Session_Timeout":
		return slp.SendINVITESessionTimeout

	case "Send_CANCEL":
		return slp.SendCANCEL

	default:
		return nil
	}
}

func (slp *TestSLP) RecvMESSAGE(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// Send MESSAGE 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) RecvINVITE(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		slp.req = m

		// Recv INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)

		// Send INVITE 180
		resp := rtclib.JSIPMsgRes(m, 180)
		rtclib.SendMsg(resp)
	} else if slp.count == 1 {
		// Recv PRACK
		assert(m.Type == rtclib.PRACK)
		assert(m.Code == 0)

		// Send PRACK 200
		pr200 := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(pr200)

		// Send UPDATE
		update := rtclib.JSIPMsgUpdate(slp.req)
		rtclib.SendMsg(update)
	} else if slp.count == 2 {
		// Recv UPDATE 200
		assert(m.Type == rtclib.UPDATE)
		assert(m.Code == 200)

		// Send INVITE 200
		resp := rtclib.JSIPMsgRes(slp.req, 200)
		rtclib.SendMsg(resp)
	} else if slp.count == 3 {
		// Recv ACK
		assert(m.Type == rtclib.ACK)
		assert(m.Code == 0)
	} else if slp.count == 4 {
		// Recv Re-INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)

		// Send INVITE 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)
	} else if slp.count == 5 { // No ACK is Ignore
		// Recv BYE
		assert(m.Type == rtclib.BYE)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) RecvINVITETimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		slp.req = m

		// Recv INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)
	} else if slp.count == 1 {
		// Recv CANCEL
		assert(m.Type == rtclib.CANCEL)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) RecvMESSAGETimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		slp.req = m

		// Recv INVITE
		assert(m.Type == rtclib.MESSAGE)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) RecvINVITESessionTimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		slp.req = m

		// Recv INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)

		// Send INVITE 180
		resp180 := rtclib.JSIPMsgRes(m, 180)
		rtclib.SendMsg(resp180)

		// Send INVITE 200
		resp200 := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp200)
	} else if slp.count == 1 {
		// Recv ACK
		assert(m.Type == rtclib.ACK)
		assert(m.Code == 0)
	} else if slp.count == 2 { // No ACK is Ignore
		// Recv BYE
		assert(m.Type == rtclib.BYE)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) RecvCANCEL(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		slp.req = m

		// Recv INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)
	} else if slp.count == 1 {
		// Recv CANCEL
		assert(m.Type == rtclib.CANCEL)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++

}

func (slp *TestSLP) SendErrorMsg(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send Error Msg
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendErrorMsg)
		invite := rtclib.JSIPMsgReq(rtclib.INVITE, "", "SendErrorMsg", "", dlg)
		rtclib.SendMsg(invite)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendBYE481(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send BYE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendBYE481)
		bye := rtclib.JSIPMsgReq(rtclib.BYE, user, "SendBYE481", user, dlg)
		rtclib.SendMsg(bye)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendACKNoSession(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send ACK
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendACKNoSession)
		ack := rtclib.JSIPMsgReq(rtclib.ACK, user, "SendACKNoSession", user, dlg)
		rtclib.SendMsg(ack)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendCANCELNoSession(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send CANCEL
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendCANCELNoSession)
		cancel := rtclib.JSIPMsgReq(rtclib.CANCEL, user, "SendCANCELNoSession", user, dlg)
		rtclib.SendMsg(cancel)

		slp.task.SetFinished()
	}

	slp.count++

}

func (slp *TestSLP) SendRespNoSession(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send Response
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendRespNoSession)
		respa := rtclib.JSIPMsgRes(m, 200)
		respa.DialogueID = dlg
		rtclib.SendMsg(respa)

		slp.task.SetFinished()
	}

	slp.count++

}

func (slp *TestSLP) SendUPDATENoSession(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send UPDATE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendUPDATENoSession)
		update := rtclib.JSIPMsgReq(rtclib.UPDATE, user, "SendUPDATENoSession", user, dlg)
		rtclib.SendMsg(update)
	} else if slp.count == 1 {
		// Recv UPDATE 481
		assert(m.Type == rtclib.UPDATE)
		assert(m.Code == 481)

		slp.task.SetFinished()
	}

	slp.count++

}

func (slp *TestSLP) SendMessage(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send MESSAGE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendMessage)
		message := rtclib.JSIPMsgReq(rtclib.MESSAGE, user, "SendMessage", user, dlg)
		rtclib.SendMsg(message)
	} else if slp.count == 1 {
		// Recv MESSAGE 200
		assert(m.Type == rtclib.MESSAGE)
		assert(m.Code == 200)

		slp.task.SetFinished()
	}

	slp.count++

}

func (slp *TestSLP) SendINVITE(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send INVITE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendINVITE)
		invite := rtclib.JSIPMsgReq(rtclib.INVITE, user, "SendINVITE", user, dlg)
		invite.SetUint("Expire", 60)
		rtclib.SendMsg(invite)
	} else if slp.count == 1 {
		// Recv INVITE 180
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 180)

		// Send PRACK
		prack := rtclib.JSIPMsgReq(rtclib.PRACK, "PRACK", "SendINVITE", "PRACK", m.DialogueID)
		rtclib.SendMsg(prack)
	} else if slp.count == 2 {
		// Recv PRACK 200
		assert(m.Type == rtclib.PRACK)
		assert(m.Code == 200)
	} else if slp.count == 3 {
		// Recv UPDATE
		assert(m.Type == rtclib.UPDATE)
		assert(m.Code == 0)

		// Send UPDATE 200
		up200 := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(up200)
	} else if slp.count == 4 {
		// Recv INVITE 200
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 200)

		// not Send ACK
	} else if slp.count == 5 {
		// Recv Re-INVITE
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 0)

		// Send Re-INVITE 200
		reinvite200 := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(reinvite200)
	} else if slp.count == 6 {
		// Recv ACK
		assert(m.Type == rtclib.ACK)
		assert(m.Code == 0)

		time.Sleep(30 * time.Second)

		// Send BYE
		bye := rtclib.JSIPMsgReq(rtclib.BYE, "BYE", "SendINVITE", "BYE", m.DialogueID)
		rtclib.SendMsg(bye)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendINVITETimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send INVITE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendINVITETimeout)
		invite := rtclib.JSIPMsgReq(rtclib.INVITE, user, "SendINVITETimeout", user, dlg)
		rtclib.SendMsg(invite)
	} else if slp.count == 1 {
		// Recv INVITE 180
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 180)
	} else if slp.count == 2 {
		// Recv INVITE 408
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 408)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendMESSAGETimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send INVITE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendMESSAGETimeout)
		invite := rtclib.JSIPMsgReq(rtclib.MESSAGE, user, "SendMESSAGETimeout", user, dlg)
		rtclib.SendMsg(invite)
	} else if slp.count == 1 {
		// Recv INVITE 408
		assert(m.Type == rtclib.MESSAGE)
		assert(m.Code == 408)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendINVITESessionTimeout(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send INVITE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendINVITESessionTimeout)
		invite := rtclib.JSIPMsgReq(rtclib.INVITE, user, "SendINVITESessionTimeout", user, dlg)
		invite.SetUint("Expire", 60)
		rtclib.SendMsg(invite)
	} else if slp.count == 1 {
		// Recv INVITE 180
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 180)
	} else if slp.count == 2 {
		// Recv INVITE 200
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 200)

		// not Send ACK
		ack := rtclib.JSIPMsgAck(m)
		rtclib.SendMsg(ack)
	} else if slp.count == 3 {
		// Recv BYE
		assert(m.Type == rtclib.BYE)
		assert(m.Code == 0)

		slp.task.SetFinished()
	}

	slp.count++
}

func (slp *TestSLP) SendCANCEL(m *rtclib.JSIP) {
	slp.task.LogError("Recv Msg %s", m.String())

	if slp.count == 0 {
		// send OPTIONS 200
		resp := rtclib.JSIPMsgRes(m, 200)
		rtclib.SendMsg(resp)

		// Send INVITE
		user := m.From
		dlg := slp.task.NewDialogueIDWithEntry(slp.SendCANCEL)
		invite := rtclib.JSIPMsgReq(rtclib.INVITE, user, "SendCANCEL", user, dlg)
		invite.SetUint("Expire", 60)
		rtclib.SendMsg(invite)

		// Send CANCEL
		cancel := rtclib.JSIPMsgCancel(invite)
		rtclib.SendMsg(cancel)
	} else if slp.count == 1 {
		// Recv 487
		assert(m.Type == rtclib.INVITE)
		assert(m.Code == 487)

		slp.task.SetFinished()
	}

	slp.count++

}
