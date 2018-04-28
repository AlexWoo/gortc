//
//
// session.go

package main

import (
    "log"

    "janus"
    "rtclib"
    "github.com/tidwall/gjson"
    simplejson "github.com/bitly/go-simplejson"
)

type session struct {
    jsipID        string
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

    log.Printf("create janus session %d success", s.sessId)
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
        s.janusRoom = janusRoom
        return
    }
    log.Printf("getRoom: can't find room for id `%s`", s.jsipRoom)

    var msg janus.ClientMsg
    j := s.janusConn

    janusSess, _ := j.Session(s.sessId)
    tid := janusSess.NewTransaction()

    msg.Janus = "message"
    msg.Transaction = tid
    msg.SessionId = s.sessId
    msg.HandleId = s.handleId
    msg.Body.Request = "create"
    msg.Body.Audiocodec = "opus"
    msg.Body.Videocodec = "h264"

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
    log.Printf("getRoom: add room `%d` for id `%s`", s.janusRoom, s.jsipRoom)

    log.Printf("getRoom: create room %d for session %d", s.janusRoom, s.sessId)
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
        RequestURI: s.jsipRoom,
        From:       feed.display,
        To:         s.userName,
        DialogueID: dialogueID,
        Body:       body,
    }

    rtclib.SendJsonSIPMsg(nil, request)
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
