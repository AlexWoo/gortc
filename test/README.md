# Test Case
---
## Prepare

1. Make sure gortc is running

		git clone https://github.com/AlexWoo/gortc.git
		cd gortc
		./install
		/usr/local/gortc/bin/gortc

	gortc will install under /usr/local/gortc, by default, rtcserver port open at 8080, apiserver port open at 2539

2. Load test slp in default

		/usr/local/gortc/bin/pcompile test
		curl -XPOST http://127.0.0.1:2539/slpm/v1/default?file=test.so

## Test case

### JSIP Unit test

	cd src/rtclib
	export GOPATH=/usr/local/gortc
	go test

### JSIP test

JSIP test script is under gortc/test/script. We use rtest for gortc signalling test

	cd test/script

- SLP not load

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f slp_not_find

- Recv BYE 481

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_bye_481

- Recv ACK no Session

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_ack_no_session

- Recv MESSAGE

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_message

- Recv INVITE

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_invite

- Recv INVITE Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_invite_timeout

- Recv MESSAGE Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_message_timeout

- Recv INVITE Session Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_invite_session_timeout

- Recv INVITE Cancel

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f recv_cancel

- Send Error Msg

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_error_msg

- Send BYE 481

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_bye_481

- Send ACK no Session

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_ack_no_session

- Send CANCEL no Session

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_cancel_no_session

- Send Resp no Session

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_resp_no_session

- Send UPDATE no Session

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_update_no_session

- Send MESSAGE

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_message

- Send INVITE

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_invite

- Send INVITE Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_invite_timeout

- Send MESSAGE Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_message_timeout

- Send INVITE Session Timeout

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_invite_session_timeout

- Send INVITE Cancel

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f send_cancel