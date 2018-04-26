//
//
// session.go

package main

import (
    "log"

    "janus"
    simplejson "github.com/bitly/go-simplejson"
)

type session struct {
    jsipID        string
    videoroom    *Videoroom
    janusConn    *janus.Janus
    mutex         chan struct{}
    sessId        int
    handleId      int
    jsipRoom      string
    janusRoom     int64
    userName      string
    myId          int64
    myPrivateId   int64
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
    log.Printf("receive from channel: %+v", req)
    j.NewSess(req.Data.Id)
    s.sessId = req.Data.Id

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
    log.Printf("receive from channel: %+v", req)
    janusSess.Attach(req.Data.Id)
    s.handleId = req.Data.Id

    log.Printf("attach handle %d for session %d", s.handleId, s.sessId)
}

func (s *session) getRoom() {
    janusRoom, exist := s.videoroom.rooms[s.jsipRoom]
    if exist {
        s.janusRoom = janusRoom
        return
    }

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
    log.Printf("receive from channel: %+v", req)
    s.janusRoom = req.Plugindata.Data.Room
    s.videoroom.rooms[s.jsipRoom] = s.janusRoom

    log.Printf("create room %d for session %d", s.janusRoom, s.sessId)
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
    for req.Janus != "event" {
        log.Printf("joinRoom: receive from channel: %+v", req)
        req = <- reqChan
    }
    log.Printf("receive from channel: %+v", req)
    s.myId = req.Plugindata.Data.Id
    s.myPrivateId = req.Plugindata.Data.PrivateId
    // TODO: new remote feeder
    // req.Plugindata.Data.Publisher

    log.Printf("join room %d for session %d", s.janusRoom, s.sessId)
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
    for req.Janus != "event" {
        log.Printf("joinRoom: receive from channel: %+v", req)
        req = <- reqChan
    }

    if req.Jsep.Type != "answer" {
        log.Printf("joinRoom: get answer failed. msg: %+v", req)
        return ""
    }

    log.Printf("receive from channel: %+v", req)
    return req.Jsep.Sdp
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
    log.Printf("candidate completed: receive from channel: %+v", req)
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
    log.Printf("candidate: receive from channel: %+v", req)
}
