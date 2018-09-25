package main

import (
	"net/http"
	"rtclib"
	"strings"
)

type apiv1 struct {
	c *ctx
}

func APIInstance() rtclib.API {
	return &apiv1{
		c: c,
	}
}

func (i *apiv1) listrooms() (int, *map[string]string, interface{},
	*map[int]rtclib.RespCode) {

	rooms := []string{}

	i.c.roomsLock.RLock()
	for k := range i.c.rooms {
		rooms = append(rooms, k)
	}
	i.c.roomsLock.RUnlock()

	ret := map[string][]string{
		"rooms": rooms,
	}

	return 0, nil, ret, nil
}

func (i *apiv1) roominfo(roomid string) (int, *map[string]string, interface{},
	*map[int]rtclib.RespCode) {

	users := make(map[string]string)

	i.c.roomsLock.RLock()
	room, ok := i.c.rooms[roomid]
	i.c.roomsLock.RUnlock()
	if ok {
		room.usersLock.RLock()
		for k, v := range room.users {
			users[k] = v.nickname
		}
		room.usersLock.RUnlock()

		ret := map[string]map[string]string{
			roomid: users,
		}

		return 0, nil, ret, nil
	}

	return 3, nil, "room " + roomid + " not exist", nil
}

func (i *apiv1) delroom(roomid string) (int, *map[string]string, interface{},
	*map[int]rtclib.RespCode) {

	i.c.roomsLock.RLock()
	room, ok := i.c.rooms[roomid]
	i.c.roomsLock.RUnlock()
	if ok {
		i.c.roomsLock.RLock()
		for _, user := range room.users {
			room.quit <- user
		}
		i.c.roomsLock.RUnlock()
	}

	return 0, nil, nil, nil
}

func (i *apiv1) deluser(roomid string, userid string) (int, *map[string]string,
	interface{}, *map[int]rtclib.RespCode) {

	i.c.roomsLock.RLock()
	room, ok := i.c.rooms[roomid]
	i.c.roomsLock.RUnlock()
	if ok {
		room.usersLock.RLock()
		user, ok1 := room.users[userid]
		room.usersLock.RUnlock()
		if ok1 {
			room.quit <- user
		}
	}

	return 0, nil, nil, nil
}

// /chatroom/v1/rooms
//	GET All rooms in chatroom
//
// /chatroom/v1/rooms/<roomid>
//	GET room info of <roomid>
func (i *apiv1) Get(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	para := strings.Split(paras, "/")

	switch para[0] {
	case "rooms":
		switch len(para) {
		case 1:
			return i.listrooms()
		case 2:
			return i.roominfo(para[1])
		default:
			return 3, nil, "Unsupported para number", nil
		}
	}

	return 3, nil, "Unknown Para", nil
}

func (i *apiv1) Post(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	return 3, nil, nil, nil
}

// /chatroom/v1/rooms/<roomid>
//	DELETE room <roomid>
//
// /chatroom/v1/rooms/<roomid>/<userid>
//	DELETE <userud> from <roomid>
func (i *apiv1) Delete(req *http.Request, paras string) (int,
	*map[string]string, interface{}, *map[int]rtclib.RespCode) {

	para := strings.Split(paras, "/")

	switch para[0] {
	case "rooms":
		switch len(para) {
		case 2:
			return i.delroom(para[1])
		case 3:
			return i.deluser(para[1], para[2])
		default:
			return 3, nil, "Unsupported para number", nil
		}
	}

	return 3, nil, "Unknown para", nil
}
