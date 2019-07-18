// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// JSIP Response Test Case

package rtclib

import (
	"fmt"
	"testing"
)

func TestJSIPRespType(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPRespType")

	if JSIPRespType(0) != JSIPReq {
		t.Error("code for JSIPReq error")
	}

	if JSIPRespType(180) != JSIPProvisionalResp {
		t.Error("code for JSIPProvisionalResp error")
	}

	if JSIPRespType(200) != JSIPSuccessResp {
		t.Error("code for JSIPSuccessResp error")
	}

	if JSIPRespType(302) != JSIPRedirectionResp {
		t.Error("code for JSIPRedirectionResp error")
	}

	if JSIPRespType(404) != JSIPClientErrResp {
		t.Error("code for JSIPClientErrResp error")
	}

	if JSIPRespType(500) != JSIPServerErrResp {
		t.Error("code for JSIPServerErrResp error")
	}

	if JSIPRespType(699) != JSIPGlobalFailureResp {
		t.Error("code for JSIPGlobalFailureResp error")
	}

	if JSIPRespType(-1) != JSIPResponseType(Unknown) {
		t.Error("error code error")
	}

	if JSIPRespType(700) != JSIPResponseType(Unknown) {
		t.Error("error code error")
	}
}

func TestJSIPRespDesc(t *testing.T) {
	fmt.Println("!!!!!!!!!!TestJSIPRespDesc")

	if JSIPRespDesc(0) != "" {
		t.Error("desc for code 0 error")
	}

	if JSIPRespDesc(-1) != "" {
		t.Error("desc for code -1 error")
	}

	if JSIPRespDesc(2000) != "" {
		t.Error("desc for code 2000 error")
	}

	if JSIPRespDesc(180) != "Ringing" {
		t.Error("desc for code 180 error")
	}

	if JSIPRespDesc(101) != "User Defined Provisional Response" {
		t.Error("desc for code 101 error")
	}

	if JSIPRespDesc(299) != "User Defined Success Response" {
		t.Error("desc for code 299 error")
	}

	if JSIPRespDesc(388) != "User Defined Redirection Response" {
		t.Error("desc for code 388 error")
	}

	if JSIPRespDesc(477) != "User Defined Client Error Response" {
		t.Error("desc for code 477 error")
	}

	if JSIPRespDesc(566) != "User Defined Server Error Response" {
		t.Error("desc for code 566 error")
	}

	if JSIPRespDesc(655) != "User Defined Global Failure Response" {
		t.Error("desc for code 655 error")
	}
}
