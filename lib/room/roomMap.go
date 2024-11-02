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
	defer roomMap.Unlock()

	room := roomMap.internal[roomId]
	if room == nil {
		room = roomMap.newRoom(roomId, name, description)
		roomMap.internal[roomId] = room
	} else if room.Name != name {
		room.Name = name
	}

	return room
}

func (roomMap *RoomMap) DeleteRoom(roomId int64) {
	roomMap.Lock()
	defer roomMap.Unlock()
	delete(roomMap.internal, roomId)
}
