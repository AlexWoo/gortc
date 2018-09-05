# SLP Process in go rtc server
---
## Receive a http request on RTC Server port

In default config, RTC server will listen on :8080, and serve location is /rtc.

When a new http request connect to RTC server, it will check para userid, if not, will return 400, Miss userid to client.

The http request must be a websocket request, otherwise, client will receive 400 response.

RTC Server user userid to ident a websocket channel, if user quit and reconnect with same userid, RTC Server will use new connection replace old connection, and send msg buffered in old failue websocket connection continued.

## Receive a msg in websocket

After a websocket connection established, rtc server will process msg from client.

When a msg received on websocket connection, go rtc server will send it go jsip stack first.

Jsip stack will check syntax first. And then, msg will send to transaction layer, then session layer. Trasaction layer and session layer will check state to judge msg is legal. Illegal msg will reject by related response msg. [The detail of jsip stack process](jsip_stack.md).

Then, jsip stack will send msg to distribute module of go rtc server. The distribute module will send msg to slp instance, if related slp instance exist. Other wise, it will create a new slp instance and send msg to slp instance. [The detail of distribute module](task.md)

## Send a msg in websocket

When SLP Send a msg, it will go through jsip stack session layer, transaction layer and syntax layer. Wrong msg will block in jsip stack.

If the msg is not first request msg for a new SIP session, jsip stack will find the websocket channel by DialogueID in jsip msg, and send the msg.

If the msg is first request msg for a new SIP session, jsip stack will get the name of next node where msg sending to. If jsip msg has Router header, jsip stack will get name from UserWithHost in first Router uri. Otherwise, it will get name from UserWithHost in RequestUri. If named websocket channel exists, multiplex websocket channel, otherwise, create a new websocket channel with the name.
