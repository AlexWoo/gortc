//
//
// janusHeap.go

package main

import (
    "janus"
)


type janusItem struct {
    janusConn   *janus.Janus
    numSess      uint64
}

type janusHeap []*janusItem

func (jh janusHeap) Len() int {
    return len(jh)
}

func (jh janusHeap) Less(i, j int) bool {
    return jh[i].numSess < jh[j].numSess
}

func (jh janusHeap) Swap(i, j int) {
    jh[i], jh[j] = jh[j], jh[i]
}

func (jh *janusHeap) Push(x interface{}) {
    item := x.(*janusItem)
    *jh = append(*jh, item)
}

func (jh *janusHeap) Pop() interface{} {
    old := *jh
    n := len(old)
    item := old[n-1]
    *jh = old[0 : n-1]
    return item
}
