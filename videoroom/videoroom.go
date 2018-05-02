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


type globalCtx struct {
    sessions    map[string](*session)
    rooms       map[string]uint64
    jh         *janusHeap
}

type Config struct {
    JanusAddr       string
    MaxConcurrent   uint64
}

type Videoroom struct {
    task       *rtclib.Task
    config     *Config
    sess       *session
    ctx        *globalCtx
    mutex       chan struct{}
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
    var vr = &Videoroom{task: task,
                        mutex: make(chan struct{}, 1)}
    if !vr.loadConfig() {
        log.Println("Videoroom load config failed")
    }

    if task.Ctx.Body == nil {
        ctx := &globalCtx{sessions: make(map[string](*session)),
                          rooms:make(map[string]uint64),
                          jh: &janusHeap{},}
        heap.Init(ctx.jh)

        task.Ctx.Body = ctx
        log.Printf("Init videoroom ctx, ctx = %+v", task.Ctx.Body)
    }

    vr.ctx = task.Ctx.Body.(*globalCtx)
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

func (vr *Videoroom) lock() {
    vr.mutex <- struct{}{}
    log.Printf("lock videoroom")
}

func (vr *Videoroom) unlock() {
    <- vr.mutex
    log.Printf("unlock videoroom")
}

func (vr *Videoroom) cachedOrNewJanus() *janus.Janus {
    if vr.ctx.jh.Len() > 0 && (*vr.ctx.jh)[0].numSess < vr.config.MaxConcurrent {
        (*vr.ctx.jh)[0].numSess += 1
        log.Printf("use a cached janus instance with numSess %d",
                   (*vr.ctx.jh)[0].numSess)
        return (*vr.ctx.jh)[0].janusConn
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

    heap.Push(vr.ctx.jh, &janusItem{janusConn: j, numSess: 1})
    log.Printf("created a new janus instance")

    return j
}

func (vr *Videoroom) setRoom(id string, room uint64) {
    vr.ctx.rooms[id] = room
}

func (vr *Videoroom) getRoom(id string) (uint64, bool) {
    room, exist := vr.ctx.rooms[id]
    return room, exist
}

func (vr *Videoroom) newSession(jsip *rtclib.JSIP) (*session, bool) {
    conn := vr.cachedOrNewJanus()
    if conn == nil {
        return nil, false
    }

    DialogueID := jsip.DialogueID
    sess := &session{jsipID: DialogueID,
                     url: jsip.RequestURI,
                     janusConn: conn,
                     videoroom: vr,
                     jsipRoom: jsip.To,
                     userName: jsip.From,
                     mutex: make(chan struct{}, 1),
                     feeds: make(map[string](*feed)),}
    vr.sess = sess
    vr.ctx.sessions[DialogueID] = sess
    log.Printf("create videoroom for DialogueID %s success", DialogueID)
    return sess, true
}

func (vr *Videoroom) setSession(id string, sess *session) {
    vr.ctx.sessions[id] = sess
}

func (vr *Videoroom) cachedSession() (*session) {
    sess := vr.sess
    return sess
}

func (ctx *globalCtx) cachedSession(DialogueID string) (*session, bool) {
    sess, exist := ctx.sessions[DialogueID]
    return sess, exist
}

func (vr *Videoroom) processINVITE(jsip *rtclib.JSIP) {
    if jsip.Code != 0 {
        vr.processFeed(jsip)
        return
    }

    vr.lock()
    sess, ok := vr.newSession(jsip)
    if !ok {
        log.Printf("invite: create session failed")
        return
    }
    sess.lock()
    vr.unlock()
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
    vr.lock()
    sess, ok := vr.ctx.cachedSession(jsip.DialogueID)
    vr.unlock()
    if !ok {
        log.Printf("not found cached session for id %s", jsip.DialogueID)
    }

    candidate, exist := jsip.Body.(*simplejson.Json).CheckGet("candidate")
    if exist == false {
        /* Now the info only transport candidate */
        log.Printf("not found candidate")
        return
    }

    feed, ok := sess.cachedFeed(jsip.DialogueID)
    if !ok {
        sess.lock()
        sess.candidate(candidate)
        sess.unlock()
    } else {
        feed.candidate(candidate)
    }

    rtclib.SendJSIPRes(jsip, 200)
}

func (vr *Videoroom) processFeed(jsip *rtclib.JSIP) {
    sess, ok := vr.ctx.cachedSession(jsip.DialogueID)
    if !ok {
        log.Printf("not found cached session for id %s", jsip.DialogueID)
    }

    feed, ok := sess.cachedFeed(jsip.DialogueID)
    if !ok {
        log.Printf("not found cached feed for id %s for sess %s",
                   jsip.DialogueID, sess.jsipID)
        return
    }

    answer, err := jsip.Body.(*simplejson.Json).Get("sdp").String()
    if err != nil {
        log.Printf("not found sdp for session %s in the respnse", sess.jsipID)
        return
    }

    feed.start(answer)

    resp := &rtclib.JSIP{
        Type:       rtclib.ACK,
        RequestURI: sess.jsipRoom,
        From:       jsip.From,
        To:         jsip.To,
        DialogueID: jsip.DialogueID,
    }

    rtclib.SendJsonSIPMsg(nil, resp)
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

