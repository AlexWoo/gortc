package main

import (
    "log"
    "container/heap"
    "time"
    "strconv"
    "net/http"
    "encoding/json"
    "sync"
    "bytes"
    "io/ioutil"

    simplejson "github.com/bitly/go-simplejson"
    "github.com/tidwall/gjson"
    "github.com/go-ini/ini"
    "github.com/go-redis/redis"
    "rtclib"
    "janus"
)


type globalCtx struct {
    sessions    map[string](*session)
    jh         *janusHeap
    sessLock    sync.RWMutex
    rClient    *redis.Client
}

type Config struct {
    JanusAddr       string
    MaxConcurrent   uint64
    RedisAddr       string
    RedisPassword   string
    RedisDB         uint64
    ApiAddr         string
}

type Videoroom struct {
    task       *rtclib.Task
    config     *Config
    sess       *session
    ctx        *globalCtx
    handleFlag  bool
    msgChan     chan *rtclib.JSIP
    mutex       chan struct{}
    register    bool
}

func GetInstance(task *rtclib.Task) rtclib.SLP {
    var vr = &Videoroom{task: task,
                        msgChan: make(chan *rtclib.JSIP, 5),
                        mutex: make(chan struct{}, 1)}
    if !vr.loadConfig() {
        log.Println("Videoroom load config failed")
    }

    if task.Ctx.Body == nil {
        ctx := &globalCtx{sessions: make(map[string](*session)),
                          jh: &janusHeap{},}
        heap.Init(ctx.jh)

        ctx.rClient = redis.NewClient(
            &redis.Options{
                Addr:     vr.config.RedisAddr,
                Password: vr.config.RedisPassword,
                DB:       int(vr.config.RedisDB),
            })
        pong, err := ctx.rClient.Ping().Result()
        if err != nil {
            log.Println("Init videoroom: ping redis err: ", err)
        }
        log.Println("Init videoroom: ping redis, response: ", pong)

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

func (vr *Videoroom) incrMember(room string) {
    err := vr.ctx.rClient.HIncrBy("member", room, 1).Err()
    if err != nil {
        log.Printf("incrMember: redis err: %s", err.Error())
    }
}

func (vr *Videoroom) decrMember(room string) {
    num, err := vr.ctx.rClient.HIncrBy("member", room, -1).Result()
    if err != nil {
        log.Printf("decrMember: redis err: %s", err.Error())
        return
    }

    if num <= 0 {
        log.Printf("decrMember: room %s don't have any member, delete it", room)
        vr.delMember(room)
        vr.delRoom(room)
        if vr.register {
            vr.deregisterRoom(vr.sess.jsipRoom)
        }
    }
}

func (vr *Videoroom) delMember(room string) {
    num, err := vr.ctx.rClient.HDel("member", room).Result()
    if err != nil {
        log.Printf("delMember: redis err: %s", err.Error())
    }

    log.Printf("delMember: delete %d for room %s", num, room)
}

func (vr *Videoroom) registerRoom(id string, room uint64) {
    msg := make(map[string]interface{})
    msg["janus"] = vr.config.JanusAddr
    msg["room"] = room

    b, err := json.Marshal(msg)
    if err != nil {
        log.Println("registerRoom: json err: ", err)
        return
    }

    addr := vr.config.ApiAddr + "/register/" + id
    contentType := "application/json;charset=utf-8"
    body := bytes.NewBuffer(b)

    resp, err := http.Post(addr, contentType, body)
    if err != nil {
        log.Println("registerRoom: post err: ", err)
        return
    }

    defer resp.Body.Close()
    rBody, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("registerRoom: read err: ", err)
    }

    if resp.StatusCode != http.StatusOK {
        log.Printf("registerRoom: register failed, response code: `%d`, " +
                   "body: `%s`", resp.StatusCode, rBody)
        return
    }

    if gjson.GetBytes(rBody, "register").Bool() == true {
        vr.register = true
        return
    }

    janusAddr := gjson.GetBytes(rBody, "janus").String()
    remoteRoom := gjson.GetBytes(rBody, "room").Uint()
    vr.sess.route(janusAddr, remoteRoom)
}

func (vr *Videoroom) deregisterRoom(id string) {
    addr := vr.config.ApiAddr + "/deregister/" + id
    resp, err := http.Get(addr)
    if err != nil {
        log.Println("deregisterRoom: get err: ", err)
        return
    }

    if resp.StatusCode != http.StatusOK {
        log.Printf("deregisterRoom: deregister failed, response code: `%d`",
                   resp.StatusCode)
        return
    }

    vr.register = false
    return
}

func (vr *Videoroom) setRoom(id string, room uint64) {
    err := vr.ctx.rClient.HSet("room", id, room).Err()
    if err != nil {
        log.Printf("setRoom: redis err: %s", err.Error())
        return
    }

    vr.incrMember(id)

    go vr.registerRoom(id, room)
}

func (vr *Videoroom) getRoom(id string) (uint64, bool) {
    room, err := vr.ctx.rClient.HGet("room", id).Uint64()
    if err == redis.Nil {
        log.Printf("getRoom: room %s isn't existed", id)
    } else if err != nil {
        log.Printf("getRoom: redis err: %s", err.Error())
    } else {
        return room, true
    }

    return room, false
}

func (vr *Videoroom) delRoom(id string) {
    num, err := vr.ctx.rClient.HDel("room", id).Result()
    if err != nil {
        log.Printf("delRoom: redis err: %s", err.Error())
    }

    log.Printf("delRoom: delete %d for room %s", num, id)
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
    vr.setSession(DialogueID, sess)
    log.Printf("create videoroom for DialogueID %s success", DialogueID)
    return sess, true
}

func (vr *Videoroom) setSession(id string, sess *session) {
    vr.ctx.setSession(id, sess)
}

func (vr *Videoroom) cachedSession() (*session) {
    sess := vr.sess
    return sess
}

func (ctx *globalCtx) cachedSession(DialogueID string) (*session, bool) {
    ctx.sessLock.RLock()
    sess, exist := ctx.sessions[DialogueID]
    ctx.sessLock.RUnlock()
    return sess, exist
}

func (ctx *globalCtx) setSession(id string, sess *session) {
    ctx.sessLock.Lock()
    ctx.sessions[id] = sess
    ctx.sessLock.Unlock()
}

func (ctx *globalCtx) delSession(id string) {
    ctx.sessLock.Lock()
    defer ctx.sessLock.Unlock()
    delete(ctx.sessions, id)
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
        return
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
        return
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
        RawMsg:     make(map[string]interface{}),
    }

    resp.RawMsg["RelatedID"] = json.Number(strconv.FormatUint(jsip.CSeq, 10))

    rtclib.SendJsonSIPMsg(nil, resp)
}

func (vr *Videoroom) processBYE(jsip *rtclib.JSIP) {
    sess, ok := vr.ctx.cachedSession(jsip.DialogueID)
    if !ok {
        log.Printf("BYE: not found cached session for id %s", jsip.DialogueID)
        return
    }

    if sess.jsipID == jsip.DialogueID {
        sess.unpublish()
        vr.decrMember(sess.jsipRoom)
    } else {
        feed, ok := sess.cachedFeed(jsip.DialogueID)
        if !ok {
            log.Printf("BYE: not found cached feed for id %s for sess %s",
                       jsip.DialogueID, sess.jsipID)
            return
        }
        sess.detach(feed.handleId)
    }
    vr.ctx.delSession(jsip.DialogueID)
}

func (vr *Videoroom) handleMsg() {
    defer func() {
        vr.handleFlag = false
    }()

    for {
        msg := <- vr.msgChan

        switch msg.Type {
        case rtclib.INVITE:
            vr.processINVITE(msg)
        case rtclib.INFO:
            vr.processINFO(msg)
        case rtclib.BYE:
            vr.processBYE(msg)
        }
    }
}

func (vr *Videoroom) Process(jsip *rtclib.JSIP) int {
    log.Println("recv msg: ", jsip)
    log.Printf("The config: %+v", vr.config)

    if vr.handleFlag == false {
        vr.handleFlag = true
        go vr.handleMsg()
    }

    vr.msgChan <- jsip

    switch jsip.Type {
    case rtclib.BYE:
        return rtclib.FINISH
    }

    return rtclib.CONTINUE
}

