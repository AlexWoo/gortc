package main

import (
	"rtclib"
)

type Slpdemo struct {
	task *rtclib.Task
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	return &Slpdemo{task: task}
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) int {
	rtclib.SendJSIPRes(jsip, 200)

	return rtclib.FINISH
}
