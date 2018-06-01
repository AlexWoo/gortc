package main

import (
	"fmt"
	"rtclib"
)

type Slpdemo struct {
	task   *rtclib.Task
	dlguas string
	dlguac string
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
	return &Slpdemo{task: task}
}

func (slp *Slpdemo) Process(jsip *rtclib.JSIP) {
	fmt.Printf("+++++++++slp: %p %s\n", slp, rtclib.JsipName(jsip))

	if jsip.Type == rtclib.INVITE && jsip.Code == 0 {
		slp.dlguas = jsip.DialogueID
		if len(jsip.Router) != 0 {
			jsip.Router = jsip.Router[1:]
		}
		//jsip.Router = []string{}
		//jsip.RequestURI = "aaa@www.test.com:10086;lr"

		slp.dlguac = slp.task.NewDialogueID()
		rtclib.SendJSIPReq(jsip, slp.dlguac)
	} else {
		dlg := jsip.DialogueID
		if dlg == slp.dlguas {
			dlg = slp.dlguac
		} else {
			dlg = slp.dlguas
		}

		fmt.Println("++++++", dlg, jsip.DialogueID)
		rtclib.SendJSIPReq(jsip, dlg)
	}

	if jsip.Type == rtclib.BYE {
		slp.task.SetFinished()
	}

	return
}
