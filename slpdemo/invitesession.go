package main

import (
	"fmt"
	"rtclib"
)

type Slpdemo struct {
	task *rtclib.Task
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	return &Slpdemo{task: task}
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) int {
	fmt.Println("recv msg: ", jsip)

	switch jsip.Type {
	case rtclib.INVITE:
		rtclib.SendJSIPRes(jsip, 200)
	case rtclib.ACK:
		rtclib.SendJSIPBye(rtclib.Jsessions[jsip.DialogueID])
	}

	return rtclib.CONTINUE
}
