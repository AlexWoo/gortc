// Copyright (C) AlexWoo(Wu Jie) wj19840501@gmail.com
//

// ChatRoom demo
// chatroom api

package main

import (
	"net/http"
	"rtclib"
	"strings"
)

type apiv1 struct {
	manager *roomManager
}

func APIInstance() rtclib.API {
	return &apiv1{
		manager: manager,
	}
}

func (api *apiv1) listrooms() (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	api.manager.roomsLock.RLock()
	defer api.manager.roomsLock.RUnlock()

	ret := map[string][]string{
		"rooms": make([]string, len(api.manager.rooms)),
	}

	i := 0
	for id := range api.manager.rooms {
		ret["rooms"][i] = id
		i++
	}

	return 0, nil, ret, nil
}

func (api *apiv1) roominfo(roomid string) (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	api.manager.roomsLock.RLock()
	defer api.manager.roomsLock.RUnlock()

	ret := map[string]map[string]string{
		roomid: make(map[string]string),
	}

	if room, ok := api.manager.rooms[roomid]; ok {
		for id, user := range room.users {
			ret[roomid][id] = user.nickname
		}
	}

	return 0, nil, ret, nil
}

func (api *apiv1) deluser(roomid string, userid string) (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	api.manager.roomsLock.RLock()
	defer api.manager.roomsLock.RUnlock()

	if room, ok := api.manager.rooms[roomid]; ok {
		room.usersLock.RLock()
		defer room.usersLock.RUnlock()

		if user, ok := room.users[userid]; ok {
			user.subscribe(0)
		}
	}

	return 0, nil, nil, nil
}

// /chatroom/v1/rooms
//	GET All rooms in chatroom
//
// /chatroom/v1/rooms/<roomid>
//	GET room info of <roomid>
func (api *apiv1) Get(req *http.Request, paras string) (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	para := strings.Split(paras, "/")

	switch para[0] {
	case "rooms":
		switch len(para) {
		case 1:
			return api.listrooms()
		case 2:
			return api.roominfo(para[1])
		default:
			return 3, nil, "Unsupported para number", nil
		}
	}

	return 3, nil, "Unknown Para", nil
}

func (api *apiv1) Post(req *http.Request, paras string) (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	return 3, nil, nil, nil
}

// /chatroom/v1/rooms/<roomid>/<userid>
//	DELETE <userud> from <roomid>
func (api *apiv1) Delete(req *http.Request, paras string) (int, *map[string]string, interface{}, *map[int]rtclib.RespCode) {
	para := strings.Split(paras, "/")

	switch para[0] {
	case "rooms":
		switch len(para) {
		case 3:
			return api.deluser(para[1], para[2])
		default:
			return 3, nil, "Unsupported para number", nil
		}
	}

	return 3, nil, "Unknown para", nil
}
