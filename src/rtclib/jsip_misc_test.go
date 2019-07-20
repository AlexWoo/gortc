// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP misc Test Case

package rtclib

import (
	"fmt"
	"testing"
)

var m = map[string]interface{}{
	"t1": float64(99999999999),
	"t2": float64(-99999999999),
	"t3": "Test",
	"t4": map[string]string{
		"aaa": "bbb",
	},
	"t5": []interface{}{
		"aaa",
		"bbb",
		"ccc",
	},
}

type check struct {
	msg     *JSIP
	typ     JSIPType
	code    int
	recv    bool
	ignore  bool
	timeout int
}

func assert(b bool) {
	if !b {
		panic("assert failed")
	}
}

func TestGetJson(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestGetJson")

	m2 := copyBody(m)
	fmt.Println(m2)

	m3 := copyMap(m)
	fmt.Println(m3)

	m4 := copySlice(m["t5"].([]interface{}))
	fmt.Println(m4)
}

func TestCopy(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestCopy")

	i1, ok := getJsonInt(m, "t1")
	if !ok {
		t.Error("getJsonInt failed")
	}
	if i1 != 99999999999 {
		t.Error("getJsonInt value unexpected", i1)
	}

	i2, ok := getJsonInt(m, "t2")
	if !ok {
		t.Error("getJsonInt failed")
	}
	if i2 != -99999999999 {
		t.Error("getJsonInt value unexpected", i2)
	}

	ui1, ok := getJsonUint(m, "t1")
	if !ok {
		t.Error("getJsonUint failed")
	}
	if ui1 != 99999999999 {
		t.Error("getJsonUint value unexpected", ui1)
	}

	ui2, ok := getJsonUint(m, "t2")
	if !ok {
		t.Error("getJsonUint failed")
	}
	fmt.Println("ui2", ui2)

	i641, ok := getJsonInt64(m, "t1")
	if !ok {
		t.Error("getJsonInt64 failed")
	}
	if i641 != 99999999999 {
		t.Error("getJsonInt value unexpected", i641)
	}

	i642, ok := getJsonInt64(m, "t2")
	if !ok {
		t.Error("getJsonInt64 failed")
	}
	if i642 != -99999999999 {
		t.Error("getJsonInt64 value unexpected", i642)
	}

	ui641, ok := getJsonUint64(m, "t1")
	if !ok {
		t.Error("getJsonUint64 failed")
	}
	if ui641 != 99999999999 {
		t.Error("getJsonUint64 value unexpected", ui641)
	}

	ui642, ok := getJsonUint64(m, "t2")
	if !ok {
		t.Error("getJsonUint64 failed")
	}
	fmt.Println("ui642", ui642)

	e1, ok := getJsonInt(m, "t3")
	if ok {
		t.Error("get wrong type success")
	}
	fmt.Println("e1", e1)

	e2, ok := getJsonInt(m, "t4")
	if ok {
		t.Error("get wrong type success")
	}
	fmt.Println("e2", e2)

	e3, ok := getJsonInt(m, "dddd")
	if ok {
		t.Error("get non-exist success")
	}
	fmt.Println("e3", e3)
}
