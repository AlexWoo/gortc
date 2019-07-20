// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP URI Test Case

package rtclib

import (
	"fmt"
	"testing"
)

func TestJSIPUriPara(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPUriPara")

	p, err := NewJSIPUriPara("a=b")
	assert(err == nil)
	assert(p.Key == "a")
	assert(p.Value == "b")
	assert(p.String() == "a=b")

	p, err = NewJSIPUriPara("a")
	assert(err == nil)
	assert(p.Key == "a")
	assert(p.Value == "")
	assert(p.String() == "a")

	p, err = NewJSIPUriPara("a=b=c=d")
	assert(err == nil)
	assert(p.Key == "a")
	assert(p.Value == "b=c=d")
	assert(p.String() == "a=b=c=d")

	p, err = NewJSIPUriPara("   a=b	")
	assert(err == nil)
	assert(p.Key == "a")
	assert(p.Value == "b")
	assert(p.String() == "a=b")

	p, err = NewJSIPUriPara("")
	assert(err.Error() == "Null para")
}

func TestJSIPUriHostport(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPUriHostport")

	hp, err := NewJSIPUriHostport("test.com")
	assert(err == nil)
	assert(hp.Host == "test.com")
	assert(hp.Port == 0)
	assert(hp.String() == "test.com")

	hp, err = NewJSIPUriHostport("test.com:10086")
	assert(err == nil)
	assert(hp.Host == "test.com")
	assert(hp.Port == 10086)
	assert(hp.String() == "test.com:10086")

	hp, err = NewJSIPUriHostport("test.com:test")
	assert(err.Error() == "Port not int")

	hp, err = NewJSIPUriHostport("test.com:99999")
	assert(err.Error() == "Invalid port")
}

func TestJSIPUri(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPUri")

	u, err := NewJSIPUri("test.com")
	assert(err == nil)
	assert(u.User == "")
	assert(u.HostpartString() == "test.com")
	assert(u.HostportString() == "test.com")
	assert(u.UserHostString() == "test.com")
	assert(u.ParasString() == "")

	u, err = NewJSIPUri("test.com:10086")
	assert(err == nil)
	assert(u.User == "")
	assert(u.HostpartString() == "test.com:10086")
	assert(u.HostportString() == "test.com:10086")
	assert(u.UserHostString() == "test.com")
	assert(u.ParasString() == "")

	u, err = NewJSIPUri("alex@test.com")
	assert(err == nil)
	assert(u.User == "alex")
	assert(u.HostpartString() == "alex@test.com")
	assert(u.HostportString() == "test.com")
	assert(u.UserHostString() == "alex@test.com")
	assert(u.ParasString() == "")

	u, err = NewJSIPUri("alex@test.com:10086")
	assert(err == nil)
	assert(u.User == "alex")
	assert(u.HostpartString() == "alex@test.com:10086")
	assert(u.HostportString() == "test.com:10086")
	assert(u.UserHostString() == "alex@test.com")
	assert(u.ParasString() == "")

	u, err = NewJSIPUri("alex@test.com:10086;aa=b;cc")
	assert(err == nil)
	assert(u.User == "alex")
	assert(u.HostpartString() == "alex@test.com:10086")
	assert(u.HostportString() == "test.com:10086")
	assert(u.UserHostString() == "alex@test.com")
	assert(u.ParasString() == ";aa=b;cc")
	assert(u.Paras["aa"] == "b")
	cc, ok := u.Paras["cc"]
	assert(cc == "")
	assert(ok)

	u, err = NewJSIPUri("alex@test.com@10086")
	assert(err.Error() == "Error hostpart")

	u, err = NewJSIPUri("test.com:test")
	assert(err.Error() == "Port not int")
}
