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

func test(jsip interface{}) {
	fmt.Println("!!!!!!!test", time.Now(), jsip.(*rtclib.JSIP))
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) {
	fmt.Printf("slp: %p\n", slp)

	switch jsip.Type {
	case rtclib.INVITE:
		timer := rtclib.NewTimer(5*time.Second, test, jsip)
		fmt.Println("!!!!!!!!", time.Now(), timer)
		//timer.Reset(10 * time.Second)
		//timer.Stop()
		rtclib.SendJSIPRes(jsip, 200)
	case rtclib.ACK:
		time.Sleep(10)
		rtclib.SendJSIPBye(rtclib.Jsessions[jsip.DialogueID])
		slp.task.SetFinished()
	}

	return
}
