package rooms

import (
	users2 "goodbuzz/lib/rooms/users"
	"sync"
)

type roomMap struct {
	sync.Mutex
	internal map[int64]*Room
}

func (roomMap *roomMap) newRoom(roomId int64, name string) *Room {
	return &Room{
		roomId:       roomId,
		name:         name,
		buzzes:       make([]string, 0),
		buzzerStatus: Unlocked,
		players:      users2.NewUserMap[*users2.Player](),
		moderators:   users2.NewUserMap[*users2.Moderator](),
	}
}

func newRoomMap() roomMap {
	return roomMap{
		internal: make(map[int64]*Room),
	}
}

func (roomMap *roomMap) getOrCreateRoom(roomId int64, name string) *Room {
	roomMap.Lock()
	room := roomMap.internal[roomId]
	if room == nil {
		room = roomMap.newRoom(roomId, name)
		roomMap.internal[roomId] = room
	} else if room.name != name {
		room.name = name
	}
	roomMap.Unlock()

	return room
}
