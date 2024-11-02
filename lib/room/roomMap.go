package room

import (
	"sync"
)

type RoomMap struct {
	sync.Mutex
	internal map[int64]*Room
}

func NewRoomMap() RoomMap {
	return RoomMap{
		internal: make(map[int64]*Room),
	}
}

func (roomMap *RoomMap) GetOrCreateRoom(roomId int64, name string, description string) *Room {
	roomMap.Lock()
	room := roomMap.internal[roomId]
	if room == nil {
		room = roomMap.newRoom(roomId, name, description)
		roomMap.internal[roomId] = room
	} else if room.Name != name {
		room.Name = name
	}
	roomMap.Unlock()

	return room
}
