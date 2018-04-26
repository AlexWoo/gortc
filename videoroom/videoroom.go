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

func (vr *Videoroom) newSession(jsip *rtclib.JSIP) (*session, bool) {
    conn := vr.cachedOrNewJanus()
    if conn == nil {
        return nil, false
    }

    DialogueID := jsip.DialogueID
    vr.sessions[DialogueID] = &session{jsipID: DialogueID,
                                       janusConn: conn,
                                       videoroom: vr,
                                       jsipRoom: jsip.To,
                                       userName: jsip.From,
                                       mutex: make(chan struct{}, 1)}
    log.Printf("create videoroom for DialogueID %s success", DialogueID)
    return vr.sessions[DialogueID], true
}

func (vr *Videoroom) cachedSession(DialogueID string) (*session, bool) {
    sess, exist := vr.sessions[DialogueID]
    return sess, exist
}

func (vr *Videoroom) processINVITE(jsip *rtclib.JSIP) {
    sess, ok := vr.newSession(jsip)
    if !ok {
        log.Printf("invite: create session failed")
        return
    }
    sess.lock()
    sess.newJanusSession()
    sess.attachVideoroom()
    sess.getRoom()
    sess.joinRoom()
    offer, _ := jsip.Body.(*simplejson.Json).String()
    answer := sess.offer(offer)
    sess.unlock()
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
}

func (vr *Videoroom) processINFO(jsip *rtclib.JSIP) {
    sess, ok := vr.cachedSession(jsip.DialogueID)
    if !ok {
        log.Printf("not found cached session for id %s", jsip.DialogueID)
    }

    candidate, exist := jsip.Body.(*simplejson.Json).CheckGet("candidate")
    if exist == false {
        /* Now the info only transport candidate */
        log.Printf("not found candidate")
        return
    }

    sess.lock()
    sess.candidate(candidate)
    sess.unlock()
    rtclib.SendJSIPRes(jsip, 200)
}

func (vr *Videoroom) Process(jsip *rtclib.JSIP) int {
    log.Println("recv msg: ", jsip)

    log.Printf("The config: %+v", vr.config)
    switch jsip.Type {
    case rtclib.INVITE:
        go vr.processINVITE(jsip)
    case rtclib.INFO:
        go vr.processINFO(jsip)
    case rtclib.ACK:
        return rtclib.CONTINUE
    }

    return rtclib.CONTINUE
}

