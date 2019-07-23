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

		/usr/local/gortc/bin/pcompile demo/chatroom
		/usr/local/gortc/bin/slpm load default chatroom
		/usr/local/gortc/bin/apim load chatroom.v1 chatroom

## Test case

After test, make chatroom resource check:

	curl http://127.0.0.1:2539/chatroom/v1/rooms

It will list all rooms in system

- message no pai

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f message_no_pai

- subscribe no pai

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f subscribe_no_pai

- subscribe no expire

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f subscribe_no_expire

- message no room

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f message_no_room

- message no user

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f message_no_user

- unsubscribe no room

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f unsubscribe_no_room

- unsubscribe no user

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f unsubscribe_no_user

- subscribe unsubscribe

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f subscribe_unsubscribe

- subscribe expire

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f subscribe_expire

- usera to userb

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Bob@gortc.com -f userb
		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f usera

- usera message with router

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Bob@gortc.com -f userb_message_with_router
		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f usera_message_with_router

- userb no response

		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Bob@gortc.com -f userb_no_resp
		rtest -t uac -u ws://127.0.0.1:8080/rtc?userid=Alex@gortc.com -f usera_userb_no_resp