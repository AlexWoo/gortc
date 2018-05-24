//
//
// session.go

package main

import (
    "log"
    "time"

    "janus"
    "rtclib"
    "github.com/tidwall/gjson"
    simplejson "github.com/bitly/go-simplejson"
)

type session struct {
    jsipID        string
    url           string
    videoroom    *Videoroom
    janusConn    *janus.Janus
    mutex         chan struct{}
    feeds         map[string](*feed)
    sessId        uint64
    handleId      uint64
    jsipRoom      string
    janusRoom     uint64
    userName      string
    myId          uint64
    myPrivateId   uint64
}


func (s *session) lock() {
    s.mutex <- struct{}{}
    log.Printf("lock session %s", s.jsipID)
}

func (s *session) unlock() {
    <- s.mutex
    log.Printf("unlock session %d", s.jsipID)
}

func (s *session) newJanusSession() {
    var msg janus.ClientMsg
    j := s.janusConn

    tid := j.NewTransaction()

    msg.Janus = "create"
    msg.Transaction = tid

    log.Printf("create new Janus session. msg: %+v", msg)

    j.Send(msg)
    reqChan, ok := j.MsgChan(tid)
    if !ok {
        log.Printf("create: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("receive from channel: %s", req)
    s.sessId = gjson.GetBytes(req, "data.id").Uint()
    j.NewSess(s.sessId)

    go s.handleDefaultMsg()

    log.Printf("create janus session %d success", s.sessId)
}

func (s *session) handleDefaultMsg() {
    janusSess, _ := s.janusConn.Session(s.sessId)
    msgChan := janusSess.DefaultMsgChan()

    for {
        msg := <- msgChan
        switch gjson.GetBytes(msg, "janus").String() {
        case "webrtcup":
            log.Printf("webrtcup: stream(`%d`:`%d`) is up",
                       gjson.GetBytes(msg, "session_id").Uint(),
                       gjson.GetBytes(msg, "sender").Uint())
        case "hangup":
            log.Printf("hangup: stream(`%d`:`%d`) is hangup because `%s`",
                       gjson.GetBytes(msg, "session_id").Uint(),
                       gjson.GetBytes(msg, "sender").Uint(),
                       gjson.GetBytes(msg, "reason").String())
        case "event":
            pluginData := gjson.GetBytes(msg, "plugindata")
            if !pluginData.Exists() {
                log.Printf("event: no plugindata for msg `%s`", msg)
                break
            }
            data := pluginData.Get("data")
            if !data.Exists() {
                log.Printf("event: no data for msg `%s`", msg)
                break
            }
            switch data.Get("videoroom").String() {
            case "event":
                if data.Get("publishers").Exists() {
                    publishers := data.Get("publishers").Array()
                    for _, publisher := range publishers {
                        go s.listen(publisher.String())
                    }
                } else if data.Get("unpublished").Exists() {
                    unpublished := data.Get("unpublished").Uint()
                    for id, feed := range s.feeds {
                        if feed.id == unpublished {
                            s.detach(feed.handleId)

                            request := &rtclib.JSIP{
                                Type:       rtclib.BYE,
                                RequestURI: s.url,
                                From:       s.userName,
                                To:         s.jsipRoom,
                                DialogueID: id,
                            }
                            rtclib.SendJsonSIPMsg(nil, request)
                            break
                        }
                    }
                }
            }
        case "timeout":
            // TODO: Destory session
            log.Printf("timeout: session `%d` is timeout in janus server",
                       gjson.GetBytes(msg, "session_id").Uint())
            return
        }
    }
}

func (s *session) attachVideoroom() {
    var msg janus.ClientMsg
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    msg.Janus = "attach"
    msg.Plugin = "janus.plugin.videoroom"
    msg.Transaction = tid
    msg.SessionId = s.sessId

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("attach: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("receive from channel: %s", req)
    s.handleId = gjson.GetBytes(req, "data.id").Uint()
    janusSess.Attach(s.handleId)

    log.Printf("attach handle %d for session %d", s.handleId, s.sessId)
}

func (s *session) getRoom() {
    janusRoom, exist := s.videoroom.getRoom(s.jsipRoom)
    if exist {
        s.videoroom.incrMember(s.jsipRoom)
        s.janusRoom = janusRoom
        return
    }
    log.Printf("getRoom: can't find room for id `%s`", s.jsipRoom)

    j := s.janusConn
    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    body["request"] = "create"
    body["audiocodec"] = "opus"
    body["videocodec"] = "h264"
    // Don't need limit now, so just set to 20
    body["publishers"] = 20
    msg["janus"] = "message"
    msg["session_id"] = s.sessId
    msg["handle_id"] = s.handleId
    msg["transaction"] = tid
    msg["body"] = body

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("getRoom: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("getRoom: receive from channel: %s", req)
    s.janusRoom = gjson.GetBytes(req, "plugindata.data.room").Uint()
    s.videoroom.setRoom(s.jsipRoom, s.janusRoom)
    log.Printf("getRoom: add room `%d` for id `%d`", s.janusRoom, s.sessId)

    log.Printf("getRoom: create room `%d` for session `%s`", s.janusRoom,
               s.jsipRoom)
}

func (s *session) detach(hid uint64) {
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    msg["janus"] = "detach"
    msg["transaction"] = tid
    msg["session_id"] = s.sessId
    msg["handle_id"] = hid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("detach: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    if gjson.GetBytes(req, "janus").String() != "success" {
        log.Printf("detach: failed, fail message: `%s`", req)
        return
    }
    log.Printf("detach: detach handle `%d` from `%s` success", hid, s.jsipID)
}

func (s *session) unpublish() {
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    body["request"] = "unpublish"
    msg["janus"] = "message"
    msg["body"] = body
    msg["transaction"] = tid
    msg["session_id"] = s.sessId
    msg["handle_id"] = s.handleId

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("unpublish: can't find channel for tid %s", tid)
        return
    }

loop:
    for {
        req := <- reqChan
        switch gjson.GetBytes(req, "janus").String() {
        case "ack":
            log.Printf("unpublish: receive ack")
        case "error":
            log.Printf("unpublish: recevie error msg `%s`", req)
            return
        case "event":
            log.Printf("unpublish: receive success msg `%s`", req)
            break loop
        }
    }

    log.Printf("unpublish: success unpublish (`%d`:`%d`)", s.sessId, s.handleId)
}

func (s *session) joinRoom() {
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    var msg janus.ClientMsg
    msg.Janus = "message"
    msg.Transaction = tid
    msg.SessionId = s.sessId
    msg.HandleId = s.handleId
    msg.Body.Request = "join"
    msg.Body.Room = s.janusRoom
    msg.Body.Ptype = "publisher"
    msg.Body.Display = s.userName

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("joinRoom: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    for gjson.GetBytes(req, "janus").String() != "event" {
        log.Printf("joinRoom: receive from channel: %s", req)
        req = <- reqChan
    }
    log.Printf("receive from channel: %s", req)
    s.myId = gjson.GetBytes(req, "plugindata.data.id").Uint()
    s.myPrivateId = gjson.GetBytes(req, "plugindata.data.private_id").Uint()
    // TODO: new remote feeder
    // req.Plugindata.Data.Publisher
    publishers := gjson.GetBytes(req, "plugindata.data.publishers").Array()
    for _, publisher := range publishers {
        go s.listen(publisher.String())
    }

    log.Printf("join room %d for session %d", s.janusRoom, s.sessId)
}

func (s *session) listen(publisher string) {
    id := gjson.Get(publisher, "id").Uint()
    display := gjson.Get(publisher, "display").String()
    feed := newFeed(id, display, s)

    dialogueID := s.videoroom.task.NewDialogueID()
    s.feeds[dialogueID] = feed
    s.videoroom.setSession(dialogueID, s)

    feed.attachVideoroom()
    offer := feed.listen()

    type bodyStrct struct {
        Sdp       string `json:"sdp"`
    }
    body := bodyStrct{Sdp:offer}

    request := &rtclib.JSIP{
        Type:       rtclib.INVITE,
        RequestURI: s.url,
        From:       s.userName,
        To:         s.jsipRoom,
        DialogueID: dialogueID,
        Body:       body,
    }

    rtclib.SendJSIPReq(request, dialogueID)
}

func (s *session) cachedFeed(id string) (*feed, bool) {
    feed, exist := s.feeds[id]
    return feed, exist
}

func (s *session) offer(sdp string) string {
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    type Offer struct {
        janus.ClientMsg
        Jsep janus.Jsep `json:"jsep,omitempty"`
    }

    var msg Offer
    msg.Janus = "message"
    msg.Transaction = tid
    msg.SessionId = s.sessId
    msg.HandleId = s.handleId
    msg.Body.Request = "configure"
    msg.Body.Audio = true
    msg.Body.Video = true
    msg.Jsep.Type = "offer"
    msg.Jsep.Sdp = sdp

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("joinRoom: can't find channel for tid %s", tid)
        return ""
    }

    req := <- reqChan
    for gjson.GetBytes(req, "janus").String() != "event" {
        log.Printf("joinRoom: receive from channel: %s", req)
        req = <- reqChan
    }

    if gjson.GetBytes(req, "jsep.type").String() != "answer" {
        log.Printf("joinRoom: get answer failed. msg: %s", req)
        return ""
    }

    log.Printf("receive from channel: %s", req)
    return gjson.GetBytes(req, "jsep.sdp").String()
}

func (s *session) completeCandidate() {
    type candidate struct {
        janus.ClientMsg
        Candidate struct{
            Completed    bool `json:"completed"`
        } `json:"candidate"`
    }

    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    var msg candidate
    msg.Janus = "trickle"
    msg.Candidate.Completed = true
    msg.SessionId = s.sessId
    msg.HandleId = s.handleId
    msg.Transaction = tid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("candidate: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("candidate completed: receive from channel: %s", req)
}

func (s *session) candidate(candidate interface{}) {
    completed, err := candidate.(*simplejson.Json).Get("completed").Bool()
    if err == nil && completed == true {
        s.completeCandidate()
        return
    }

    type candidateStruct struct {
        janus.ClientMsg
        Candidate struct{
            Candidate       string `json:"candidate"`
            Sdpmid          string `json:"sdpMid"`
            SdpMLineIndex   int64 `json:"sdpMLineIndex"`
        } `json:"candidate"`
    }

    candiStr, _ := candidate.(*simplejson.Json).Get("candidate").String()
    sdpmid, _ := candidate.(*simplejson.Json).Get("sdpMid").String()
    sdpMLineIndex, _ := candidate.(*simplejson.Json).Get("sdpMLineIndex").Int64()

    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    var msg candidateStruct
    msg.Janus = "trickle"
    msg.Candidate.Candidate = candiStr
    msg.Candidate.Sdpmid = sdpmid
    msg.Candidate.SdpMLineIndex = sdpMLineIndex
    msg.SessionId = s.sessId
    msg.HandleId = s.handleId
    msg.Transaction = tid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("candidate: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("candidate: receive from channel: %s", req)
}

func newJanus(addr string) *janus.Janus {
    j := janus.NewJanus(addr)
    go j.WaitMsg()

    wait := 0
    for j.ConnectStatus() == false {
        if wait >= 5 {
            log.Printf("wait %d s to connect, quit", wait)
            return nil
        }
        timer := time.NewTimer(time.Second * 1)
        <- timer.C
        wait += 1
        log.Printf("wait %d s to connect", wait)
    }

    return j
}

func (s *session) route(remoteServer string, remoteRoom uint64) {
    router := newRouter(s.janusRoom, remoteRoom)
    router.remoteConn = newJanus(remoteServer)
    router.localConn = newJanus(s.videoroom.config.JanusAddr)

    router.newLocalSession()
    router.newRemoteSession()

    router.startLocal()
    router.startRemote()
}
