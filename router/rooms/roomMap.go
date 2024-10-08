package rooms

import (
	"goodbuzz/router/rooms/users"
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
		buzzerStatus: Unlocked,
		players:      users.NewUserMap[*users.Player](),
		moderators:   users.NewUserMap[*users.Moderator](),
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
