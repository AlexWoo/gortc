//
//
// router.go

package main

import (
    "log"
    "strings"

    "janus"
    "github.com/tidwall/gjson"
)

type router struct {
    localSid    uint64
    remoteSid   uint64
    localConn  *janus.Janus
    remoteConn *janus.Janus
    localRoom   uint64
    remoteRoom  uint64
    localHid    uint64
    remoteHid   uint64
    localPrivateId uint64
    remotePrivateId uint64
    listeners   map[uint64]*listener
    publishers  map[uint64]bool
}

type listener struct {
    id          uint64
    listenHid  uint64
    publishHid  uint64
}

func newListener(id uint64, listen uint64, publish uint64) *listener {
    return &listener{id: id,
                     listenHid: listen,
                     publishHid: publish,}
}

func newRouter(localRoom uint64, remoteRoom uint64) *router {
    return &router{localRoom: localRoom,
                   listeners: make(map[uint64]*listener),
                   publishers: make(map[uint64]bool),
                   remoteRoom: remoteRoom,}
}

func newJanusSession(j *janus.Janus) uint64 {
    tid := j.NewTransaction()

    msg:= make(map[string]interface{})
    msg["janus"] = "create"
    msg["transaction"] = tid

    j.Send(msg)
    reqChan, ok := j.MsgChan(tid)
    if !ok {
        log.Printf("newJanusSession: can't find channel for tid %s", tid)
    }

    req := <- reqChan
    if gjson.GetBytes(req, "janus").String() != "success" {
        log.Printf("newJanusSession: failed, fail message: `%s`", req)
        return 0
    }

    sessId := gjson.GetBytes(req, "data.id").Uint()
    j.NewSess(sessId)

    log.Printf("newJanusSession: new janus session `%d` success", sessId)
    return sessId
}

func attachVideoroom(j *janus.Janus, sid uint64) uint64 {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    msg["janus"] = "attach"
    msg["plugin"] = "janus.plugin.videoroom"
    msg["transaction"] = tid
    msg["session_id"] = sid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("attachVideoroom: can't find channel for tid %s", tid)
    }

    req := <- reqChan
    if gjson.GetBytes(req, "janus").String() != "success" {
        log.Printf("attachVideoroom: failed, fail message: `%s`", req)
        return 0
    }

    handleId := gjson.GetBytes(req, "data.id").Uint()
    janusSess.Attach(handleId)

    log.Printf("attachVideoroom: session `%d` attach handler `%d` success",
               sid, handleId)
    return handleId
}

func publish(j *janus.Janus, sid uint64, hid uint64, room uint64,
             display string) []byte {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    body["request"] = "join"
    body["room"] = room
    body["ptype"] = "publisher"
    body["display"] = display
    msg["janus"] = "message"
    msg["transaction"] = tid
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["body"] = body

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("publish: can't find channel for tid %s", tid)
    }

    for {
        req := <- reqChan
        switch gjson.GetBytes(req, "janus").String() {
        case "ack":
            log.Printf("publish: receive ack")
        case "error":
            log.Printf("publish: receive error msg `%s`", req)
            return nil
        case "event":
            log.Printf("publish: receive success msg `%s`", req)
            return req
        }
    }
}

func unpublish(j *janus.Janus, sid uint64, hid uint64) {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    body["request"] = "unpublish"
    msg["janus"] = "message"
    msg["transaction"] = tid
    msg["session_id"] = sid
    msg["handle_id"] = hid

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("unpublish: can't find channel for tid %s", tid)
    }

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
            return
        }
    }
}

func listen(j *janus.Janus, sid uint64, hid uint64, room uint64, feed uint64,
    privateId uint64) string {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    body["request"] = "join"
    body["room"] = room
    body["ptype"] = "listener"
    body["feed"] = feed
    body["private_id"] = privateId
    msg["janus"] = "message"
    msg["transaction"] = tid
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["body"] = body

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("listen: can't find channel for tid %s", tid)
    }

    for {
        req := <-reqChan
        switch gjson.GetBytes(req, "janus").String() {
        case "ack":
            log.Printf("listen: receive ack")
        case "error":
            log.Printf("listen: receive err msg `%s`", req)
            return ""
        case "event":
            log.Printf("listen: receive success msg `%s`", req)
            return gjson.GetBytes(req, "jsep.sdp").String()
        }
    }
}

func configure(j *janus.Janus, sid uint64, hid uint64, sdp string) string {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    jsep := make(map[string]interface{})
    body["request"] = "configure"
    body["audio"] = true
    body["video"] = true
    jsep["type"] = "offer"
    jsep["sdp"] = sdp
    msg["janus"] = "message"
    msg["transaction"] = tid
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["body"] = body
    msg["jsep"] = jsep

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("configure: can't find channel for tid %s", tid)
    }

    for {
        req := <- reqChan
        switch gjson.GetBytes(req, "janus").String() {
        case "ack":
            log.Printf("configure: receive ack")
        case "error":
            log.Printf("configure: receive err msg `%s`", req)
            return ""
        case "event":
            log.Printf("configure: receive success msg `%s`", req)
            return gjson.GetBytes(req, "jsep.sdp").String()
        }
    }
}

func candidate(j *janus.Janus, sid uint64, hid uint64, candiStr string,
               sdpMid string, sdpMLineIndex uint64) {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    candidate := make(map[string]interface{})
    candidate["candidate"] = candiStr
    candidate["sdpMid"] = sdpMid
    candidate["sdpMLineIndex"] = sdpMLineIndex
    msg["janus"] = "trickle"
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["transaction"] = tid
    msg["candidate"] = candidate

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("candidate: can't find channel for tid %s", tid)
    }

    req := <- reqChan
    log.Printf("candidate: receive from channel: %s", req)
}

func completeCandidate(j *janus.Janus, sid uint64, hid uint64) {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    candidate := make(map[string]interface{})
    candidate["completed"] = true
    msg["janus"] = "trickle"
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["transaction"] = tid
    msg["candidate"] = candidate

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("completeCandidate: can't find channel for tid %s", tid)
    }

    req := <- reqChan
    log.Printf("completeCandidate: receive from channel: %s", req)
}

func start(j *janus.Janus, sid uint64, hid uint64, room uint64, sdp string) {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    body := make(map[string]interface{})
    jsep := make(map[string]interface{})
    body["request"] = "start"
    body["room"] = room
    jsep["type"] = "answer"
    jsep["sdp"] = sdp
    msg["janus"] = "message"
    msg["transaction"] = tid
    msg["session_id"] = sid
    msg["handle_id"] = hid
    msg["body"] = body
    msg["jsep"] = jsep

    j.Send(msg)
    reqChan, ok := janusSess.MsgChan(tid)
    if !ok {
        log.Printf("start: can't find channel for tid %s", tid)
    }

    for {
        req := <- reqChan
        switch gjson.GetBytes(req, "janus").String() {
        case "ack":
            log.Printf("start: receive ack")
        case "error":
            log.Printf("start: receive err msg `%s`", req)
            return
        case "event":
            log.Printf("start: receive success msg `%s`", req)
            return
        }
    }
}

func detach(j *janus.Janus, sid uint64, hid uint64) {
    janusSess, _ := j.Session(sid)
    tid := janusSess.NewTransaction()

    msg := make(map[string]interface{})
    msg["janus"] = "detach"
    msg["transaction"] = tid
    msg["session_id"] = sid
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
    log.Printf("detach: detach handle `%d` from `%d` success", hid, sid)
}

func getFirstCandidateFromSdp(sdp string) string {
    for _, sdpItem := range strings.Split(sdp, "\r\n") {
        pair := strings.SplitN(sdpItem, "=", 2)
        if (len(pair) != 2) {
            continue
        }

        valuePair := strings.SplitN(pair[1], ":", 2)
        if len(valuePair) != 2{
            continue
        }

        if valuePair[0] != "candidate" {
            continue
        }

        log.Printf(
            "getFirstCandidateFromSdp: find candidate `%s` from line `%s`",
            valuePair[1], sdpItem)
        return valuePair[1]
    }

    log.Printf("getFirstCandidateFromSdp: not found candidate from sdp `%s`",
               sdp)
    return ""
}

func (r *router) handleDefaultMsg(j *janus.Janus, sid uint64) {
    janusSess, _ := j.Session(sid)
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
            sender := gjson.GetBytes(msg, "sender").Uint()
            if sender != r.localHid && sender != r.remoteHid {
                log.Printf("event: ignore router msg `%s`", msg)
                return
            }
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
                        if r.getPublisher(publisher.Get("id").Uint()) {
                            continue
                        }
                        switch sender {
                        case r.localHid:
                            go r.listenLocal(publisher.String())
                        case r.remoteHid:
                            go r.listenRemote(publisher.String())
                        }
                    }
                } else if data.Get("unpublished").Exists() {
                    unpublished := data.Get("unpublished").Uint()
                    if r.getPublisher(unpublished) {
                        r.delPublisher(unpublished)
                    }
                    listener := r.getListener(unpublished)
                    if listener == nil {
                        log.Printf("msg unpublish: id `%d` is not registered",
                                   unpublished)
                        break
                    }
                    if sender == r.localHid {
                        detach(r.localConn, r.localSid, listener.listenHid)
                        unpublish(r.remoteConn, r.remoteSid,
                                  listener.publishHid)
                        r.delListener(unpublished)
                        break
                    } else if sender == r.remoteHid {
                        detach(r.remoteConn, r.remoteSid, listener.listenHid)
                        unpublish(r.localConn, r.localSid, listener.publishHid)
                        r.delListener(unpublished)
                        break
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

func (r *router) newRemoteSession() {
    r.remoteSid = newJanusSession(r.remoteConn)
    r.remoteHid = attachVideoroom(r.remoteConn, r.remoteSid)

    go r.handleDefaultMsg(r.remoteConn, r.remoteSid)
    log.Printf("newRemoteSession: new janus session `%d` success", r.remoteSid)
}

func (r *router) newLocalSession() {
    r.localSid = newJanusSession(r.localConn)
    r.localHid = attachVideoroom(r.localConn, r.localSid)

    go r.handleDefaultMsg(r.localConn, r.localSid)
    log.Printf("newLocalSession: new janus session `%d` success", r.localSid)
}

func (r *router) registerListener(id uint64, lHid uint64, pHid uint64) {
    _, exist := r.listeners[id]
    if exist {
        log.Printf("registerListener: id `%d` already registered", id)
        return
    }
    r.listeners[id] = newListener(id, lHid, pHid)
    log.Printf("registerListener: id `%d` register success", id)
}

func (r *router) getListener(id uint64) *listener {
    listener, exist := r.listeners[id]
    if !exist {
        return nil
    }
    return listener
}

func (r *router) delListener(id uint64) {
    delete(r.listeners, id)
}

func (r *router) registerPublisher(id uint64) {
    _, exist := r.publishers[id]
    if exist {
        log.Printf("registerPublisher: id `%d` already registered", id)
        return
    }
    r.publishers[id] = true
    log.Printf("registerPublisher: id `%d` register success")
}

func (r *router) getPublisher(id uint64) bool {
    _, exist := r.publishers[id]
    // if publisher exist, it's value only will be true, so just return exist
    return exist
}

func (r *router) delPublisher(id uint64) {
    delete(r.publishers, id)
}

func (r *router) listenRemote(publisher string) {
    remoteHid := attachVideoroom(r.remoteConn, r.remoteSid)
    id := gjson.Get(publisher, "id").Uint()
    display := gjson.Get(publisher, "display").String()

    offer := listen(r.remoteConn, r.remoteSid, remoteHid, r.remoteRoom, id,
                    r.remotePrivateId)
    remoteCandidate := getFirstCandidateFromSdp(offer)

    localHid := attachVideoroom(r.localConn, r.localSid)
    publisherMsg := publish(
        r.localConn, r.localSid, localHid, r.localRoom, display)
    r.registerPublisher(
        gjson.GetBytes(publisherMsg, "plugindata.data.id").Uint())
    answer := configure(r.localConn, r.localSid, localHid, offer)
    localCandidate := getFirstCandidateFromSdp(answer)

    start(r.remoteConn, r.remoteSid, remoteHid, r.remoteRoom, answer)
    r.registerListener(id, remoteHid, localHid)

    if remoteCandidate != "" {
        candidate(r.localConn, r.localSid, localHid, remoteCandidate, "router",
                  0)
        completeCandidate(r.localConn, r.localSid, localHid)
    }

    if localCandidate != "" {
        candidate(r.remoteConn, r.remoteSid, remoteHid, localCandidate,
                  "router", 0)
        completeCandidate(r.remoteConn, r.remoteSid, remoteHid)
    }
}

func (r *router) listenLocal(publisher string) {
    localHid := attachVideoroom(r.localConn, r.localSid)
    id := gjson.Get(publisher, "id").Uint()
    display := gjson.Get(publisher, "display").String()

    offer := listen(r.localConn, r.localSid, localHid, r.localRoom, id,
                    r.localPrivateId)
    localCandidate := getFirstCandidateFromSdp(offer)

    remoteHid := attachVideoroom(r.remoteConn, r.remoteSid)
    publisherMsg := publish(
        r.remoteConn, r.remoteSid, remoteHid, r.remoteRoom, display)
    r.registerPublisher(
        gjson.GetBytes(publisherMsg, "plugindata.data.id").Uint())
    answer := configure(r.remoteConn, r.remoteSid, remoteHid, offer)
    remoteCandidate := getFirstCandidateFromSdp(answer)

    start(r.localConn, r.localSid, localHid, r.localRoom, answer)
    r.registerListener(id, localHid, remoteHid)

    if remoteCandidate != "" {
        candidate(r.localConn, r.localSid, localHid, remoteCandidate, "router",
                  0)
        completeCandidate(r.localConn, r.localSid, localHid)
    }

    if localCandidate != "" {
        candidate(r.remoteConn, r.remoteSid, remoteHid, localCandidate,
                  "router", 0)
        completeCandidate(r.remoteConn, r.remoteSid, remoteHid)
    }
}

func (r *router) startRemote() {
    publisherMsg := publish(r.remoteConn, r.remoteSid, r.remoteHid,
                            r.remoteRoom, "route")
    r.registerPublisher(
        gjson.GetBytes(publisherMsg, "plugindata.data.id").Uint())
    r.remotePrivateId = gjson.GetBytes(
        publisherMsg, "plugindata.data.private_id").Uint()

    publishers := gjson.GetBytes(
        publisherMsg, "plugindata.data.publishers").Array()
    for _, publisher := range publishers {
        if r.getPublisher(publisher.Get("id").Uint()) {
            continue
        }
        r.listenRemote(publisher.String())
    }
}

func (r *router) startLocal() {
    publisherMsg := publish(r.localConn, r.localSid, r.localHid, r.localRoom,
                            "route")
    r.registerPublisher(
        gjson.GetBytes(publisherMsg, "plugindata.data.id").Uint())
    r.localPrivateId = gjson.GetBytes(
        publisherMsg, "plugindata.data.private_id").Uint()

    publishers := gjson.GetBytes(
        publisherMsg, "plugindata.data.publishers").Array()
    for _, publisher := range publishers {
        if r.getPublisher(publisher.Get("id").Uint()) {
            continue
        }
        r.listenLocal(publisher.String())
    }
}
