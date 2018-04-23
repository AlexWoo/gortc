package main

import (
    "fmt"
    "container/heap"
    "time"

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
    janusConn    *janus.Janus
    sessId        int
    handleId      int
    videoroom    *Videoroom
}

type Videoroom struct {
    task       *rtclib.Task
    config     *Config
    sessions    map[string](*session)
    rooms       map[string]int
    jh         *janusHeap
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
    var vr = &Videoroom{task: task,
                        sessions: make(map[string](*session)),
                        jh: &janusHeap{}}
    if !vr.loadConfig() {
        fmt.Println("Videoroom load config failed")
    }

    heap.Init(vr.jh)

    return vr
}

func (vr *Videoroom) loadConfig() bool {
    vr.config = new(Config)

    confPath := rtclib.RTCPATH + "/conf/Videoroom.ini"

    f, err := ini.Load(confPath)
    if err != nil {
        fmt.Printf("Load config file %s error: %v", confPath, err)
        return false
    }

    return rtclib.Config(f, "Videoroom", vr.config)
}

func (vr *Videoroom) cachedOrNewJanus() *janus.Janus {
    if vr.jh.Len() > 0 && (*vr.jh)[0].numSess < vr.config.MaxConcurrent {
        (*vr.jh)[0].numSess += 1
        fmt.Printf("use a cached janus instance with numSess %d",
                   (*vr.jh)[0].numSess)
        return (*vr.jh)[0].janusConn
    }

    j := janus.NewJanus(vr.config.JanusAddr)
    fmt.Printf("connectd to server %s", vr.config.JanusAddr)
    go j.WaitMsg()

    wait := 0
    for j.ConnectStatus() == false {
        if wait >= 5 {
            fmt.Printf("wait %d s to connect, quit", wait)
            return nil
        }
        timer := time.NewTimer(time.Second * 1)
        <- timer.C
        wait += 1
        fmt.Printf("wait %d s to connect", wait)
    }

    heap.Push(vr.jh, &janusItem{janusConn: j, numSess: 1})
    fmt.Printf("created a new janus instance")

    return j
}

func (vr *Videoroom) session(DialogueID string) (*session, bool) {
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
                                       videoroom: vr}
    fmt.Printf("create videoroom for DialogueID %s success", DialogueID)
    return vr.sessions[DialogueID], true
}

func (s *session) newSession() {
    var msg janus.ClientMsg
    j := s.janusConn

    tid := j.NewTransaction()

    msg.Janus = "create"
    msg.Transaction = tid

    fmt.Printf("create new Janus session. msg: %+v", msg)

    j.Send(msg)
    reqChan, ok := j.MsgChan(tid)
    if !ok {
        fmt.Printf("create: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    fmt.Printf("receive from channel: %+v", req)
    j.NewSess(req.Data.Id)
    s.sessId = req.Data.Id

    fmt.Printf("create janus session %d success", s.sessId)
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
        fmt.Printf("attach: can't find channel for tid %s", tid)
        return
    }

    req := <- reqChan
    fmt.Printf("receive from channel: %+v", req)
    janusSess.Attach(req.Data.Id)
    s.handleId = req.Data.Id

    fmt.Printf("attach handle %d for session %d", s.handleId, s.sessId)
}


func (vr *Videoroom) Process(jsip *rtclib.JSIP) int {
    fmt.Println("recv msg: ", jsip)

    fmt.Printf("The config: %+v", vr.config)
    switch jsip.Type {
    case rtclib.INVITE:
        sess, ok := vr.session(jsip.DialogueID)
        if !ok {
            return rtclib.FINISH
        }
        sess.newSession()
        sess.attachVideoroom()
        // sess.getRoom()
        // vr.getRoom()
        // vr.joinRoom()
        // vr.offer()
        resp := &rtclib.JSIP{
            Type:       jsip.Type,
            Code:       200,
            From:       jsip.From,
            To:         jsip.To,
            CSeq:       jsip.CSeq,
            DialogueID: jsip.DialogueID,
            RawMsg:     make(map[string]interface{}),
        }

        rtclib.SendJsonSIPMsg(nil, resp)
    case rtclib.ACK:
        rtclib.SendJSIPBye(rtclib.Jsessions[jsip.DialogueID])
    }

    return rtclib.CONTINUE
}

