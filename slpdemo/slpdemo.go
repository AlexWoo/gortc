package main

import (
	"fmt"
	"rtclib"
	"time"
)

type Slpdemo struct {
	task *rtclib.Task
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	return &Slpdemo{task: task}
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) {
	fmt.Printf("slp: %p\n", slp)

	switch jsip.Type {
	case rtclib.INVITE:
		rtclib.SendJSIPRes(jsip, 200)
	case rtclib.ACK:
		time.Sleep(10)
		rtclib.SendJSIPBye(rtclib.Jsessions[jsip.DialogueID])
		slp.task.SetFinished()
	}

	return
}
