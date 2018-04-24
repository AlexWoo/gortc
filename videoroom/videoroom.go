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

