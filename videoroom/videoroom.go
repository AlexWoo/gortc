package main

import (
    "log"
    "container/heap"
    "time"

    simplejson "github.com/bitly/go-simplejson"
    "github.com/go-ini/ini"
    "rtclib"
    "janus"
)


type Config struct {
    JanusAddr       string
    MaxConcurrent   uint64
}

type session struct {
    jsipID        string
    videoroom    *Videoroom
    janusConn    *janus.Janus
    sessId        int
    handleId      int
    jsipRoom      string
    janusRoom     int64
    userName      string
    myId          int64
    myPrivateId   int64
}

type Videoroom struct {
    task       *rtclib.Task
    config     *Config
    sessions    map[string](*session)
    rooms       map[string]int64
    jh         *janusHeap
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
    var vr = &Videoroom{task: task,
                        sessions: make(map[string](*session)),
                        rooms:make(map[string]int64),
                        jh: &janusHeap{}}
    if !vr.loadConfig() {
        log.Println("Videoroom load config failed")
    }

    heap.Init(vr.jh)

    return vr
}

func (vr *Videoroom) loadConfig() bool {
    vr.config = new(Config)

    confPath := rtclib.RTCPATH + "/conf/Videoroom.ini"

    f, err := ini.Load(confPath)
    if err != nil {
        log.Printf("Load config file %s error: %v", confPath, err)
        return false
    }

    return rtclib.Config(f, "Videoroom", vr.config)
}

func (vr *Videoroom) cachedOrNewJanus() *janus.Janus {
    if vr.jh.Len() > 0 && (*vr.jh)[0].numSess < vr.config.MaxConcurrent {
        (*vr.jh)[0].numSess += 1
        log.Printf("use a cached janus instance with numSess %d",
                   (*vr.jh)[0].numSess)
        return (*vr.jh)[0].janusConn
    }

    j := janus.NewJanus(vr.config.JanusAddr)
    log.Printf("connectd to server %s", vr.config.JanusAddr)
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

    heap.Push(vr.jh, &janusItem{janusConn: j, numSess: 1})
    log.Printf("created a new janus instance")

    return j
}

func (vr *Videoroom) session(jsip *rtclib.JSIP) (*session, bool) {
    DialogueID := jsip.DialogueID
    sess, exist := vr.sessions[DialogueID]
    if exist {
        return sess, true
    }

    conn := vr.cachedOrNewJanus()
    if conn == nil {
        return nil, false
    }

    vr.sessions[DialogueID] = &session{jsipID: DialogueID,
                                       janusConn: conn,
                                       videoroom: vr,
                                       jsipRoom: jsip.To,
                                       userName: jsip.From}
    log.Printf("create videoroom for DialogueID %s success", DialogueID)
    return vr.sessions[DialogueID], true
}

func (s *session) newSession() {
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

func (vr *Videoroom) Process(jsip *rtclib.JSIP) int {
    log.Println("recv msg: ", jsip)

    log.Printf("The config: %+v", vr.config)
    switch jsip.Type {
    case rtclib.INVITE:
        sess, ok := vr.session(jsip)
        if !ok {
            return rtclib.FINISH
        }
        sess.newSession()
        sess.attachVideoroom()
        sess.getRoom()
        sess.joinRoom()
        offer, _ := jsip.Body.(*simplejson.Json).String()
        answer := sess.offer(offer)
        // vr.offer()
        resp := &rtclib.JSIP{
            Type:       jsip.Type,
            Code:       200,
            From:       jsip.From,
            To:         jsip.To,
            CSeq:       jsip.CSeq,
            DialogueID: jsip.DialogueID,
            RawMsg:     make(map[string]interface{}),
            Body:       answer,
        }

        rtclib.SendJsonSIPMsg(nil, resp)
    case rtclib.ACK:
        rtclib.SendJSIPBye(rtclib.Jsessions[jsip.DialogueID])
    }

    return rtclib.CONTINUE
}

