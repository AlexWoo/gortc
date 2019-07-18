// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Type Test Case

package rtclib

import (
	"fmt"
	"testing"
)

func TestNewJSIPType(t *testing.T) {
	fmt.Println("!!!!!!!!!!!TestNewJSIPType")

	t1 := "INVITE"
	t2 := "TTT"

	if NewJSIPType(t1) != INVITE {
		t.Error("NewJSIPType known type failed")
	}

	if NewJSIPType(t2) != JSIPType(Unknown) {
		t.Error("NewJSIPType unknown type failed")
	}
}

func TestJSIPTypeString(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPTypeString")

	t1 := ACK
	t2 := JSIPType(Unknown)
	t3 := TERM
	t4 := JSIPType(100)
	t5 := JSIPType(-100)

	if t1.String() != "ACK" {
		t.Error("ACK to String failed")
	}

	if t2.String() != "UNKNOWN" {
		t.Error("UNKNOWN to String failed")
	}

	if t3.String() != "TERM" {
		t.Error("TERM to String failed")
	}

	if t4.String() != "UNKNOWN" {
		t.Error("UNKNOWN to String failed")
	}

	if t5.String() != "UNKNOWN" {
		t.Error("UNKNOWN to String failed")
	}
}
