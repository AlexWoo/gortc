//
//
// feed.go

package main

import (
    "log"

    "janus"
    "github.com/tidwall/gjson"
    simplejson "github.com/bitly/go-simplejson"
)

type feed struct {
    sess      *session
    id         uint64
    display    string
    handleId   uint64
}

func newFeed(id uint64, display string, sess *session) *feed {
    return &feed{id: id,
                 display: display,
                 sess: sess}
}

func (f *feed) attachVideoroom() {
    j := f.sess.janusConn

    janusSess, _ := j.Session(f.sess.sessId)
    tid := janusSess.NewTransaction()

    var msg janus.ClientMsg
    msg.Janus = "attach"
    msg.Plugin = "janus.plugin.videoroom"
    msg.Transaction = tid
    msg.SessionId = f.sess.sessId

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("attach: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("receive from channel: %s", req)
    f.handleId = gjson.GetBytes(req, "data.id").Uint()
    janusSess.Attach(f.handleId)

    log.Printf("attach handle %d for feed %d for session %d",
               f.handleId, f.id, f.sess.sessId)
}

func (f *feed) listen() string {
    j := f.sess.janusConn

    janusSess, _ := j.Session(f.sess.sessId)
    tid := janusSess.NewTransaction()

    var msg janus.ClientMsg
    msg.Janus = "message"
    msg.Body.Request = "join"
    msg.Body.Room = f.sess.janusRoom
    msg.Body.Ptype = "listener"
    msg.Body.Feed = f.id
    msg.Body.PrivateId = f.sess.myPrivateId
    msg.Transaction = tid
    msg.SessionId = f.sess.sessId
    msg.HandleId = f.handleId

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("listen: can't find channel for tid %s", tid)
        return ""
    }

    req := <- reqChan
    for gjson.GetBytes(req, "janus").String() != "event" {
        log.Printf("joinRoom: receive from channel: %s", req)
        req = <- reqChan
    }
    log.Printf("receive from channel: %s", req)

    return gjson.GetBytes(req, "jsep.sdp").String()
}

func (f *feed) start(sdp string) {
    j := f.sess.janusConn

    janusSess, _ := j.Session(f.sess.sessId)
    tid := janusSess.NewTransaction()

    type clientMsg struct {
        janus.ClientMsg
        Jsep janus.Jsep `json:"jsep,omitempty"`
    }

    var msg clientMsg
    msg.Janus = "message"
    msg.Body.Request = "start"
    msg.Body.Room = f.sess.janusRoom
    msg.Transaction = tid
    msg.SessionId = f.sess.sessId
    msg.HandleId = f.handleId
    msg.Jsep.Type = "answer"
    msg.Jsep.Sdp = sdp

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("feed start: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    for gjson.GetBytes(req, "janus").String() != "event" {
        log.Printf("feed start: receive from channel: %s", req)
        req = <- reqChan
    }

    log.Printf("feed start: receive from channel: %s", req)
}

func (f *feed) completeCandidate() {
    type candidate struct {
        janus.ClientMsg
        Candidate struct{
            Completed    bool `json:"completed"`
        } `json:"candidate"`
    }

    j := f.sess.janusConn

    janusSess, _ := j.Session(f.sess.sessId)
    tid := janusSess.NewTransaction()

    var msg candidate
    msg.Janus = "trickle"
    msg.Candidate.Completed = true
    msg.SessionId = f.sess.sessId
    msg.HandleId = f.handleId
    msg.Transaction = tid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("feed candidate: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    log.Printf("feed candidate completed: receive from channel: %s", req)
}

func (f *feed) candidate(candidate interface{}) {
    completed, err := candidate.(*simplejson.Json).Get("completed").Bool()
    if err == nil && completed == true {
        f.completeCandidate()
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

    j := f.sess.janusConn

    janusSess, _ := j.Session(f.sess.sessId)
    tid := janusSess.NewTransaction()

    var msg candidateStruct
    msg.Janus = "trickle"
    msg.Candidate.Candidate = candiStr
    msg.Candidate.Sdpmid = sdpmid
    msg.Candidate.SdpMLineIndex = sdpMLineIndex
    msg.SessionId = f.sess.sessId
    msg.HandleId = f.handleId
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
