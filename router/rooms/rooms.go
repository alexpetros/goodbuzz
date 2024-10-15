package rooms

import (
	"context"
	"errors"
	"goodbuzz/lib/db"
	"goodbuzz/lib/room"
)

var openRooms = room.NewRoomMap()

func GetRoom(ctx context.Context, roomId int64) (*room.Room, error) {
	dbRoom := db.GetRoom(ctx, roomId)
	if dbRoom == nil {
		return nil, errors.New("Room not found")
	}

	return openRooms.GetOrCreateRoom(dbRoom.Id(), dbRoom.Name()), nil
}

func GetRoomsForTournament(ctx context.Context, tournamentId int64) []room.Room {
	dbRooms := db.GetRoomsForTournament(ctx, tournamentId)
	rooms := make([]room.Room, 0)
	for _, dbRoom := range dbRooms {
		newRoom, notFoundErr := GetRoom(ctx, dbRoom.Id())
		if notFoundErr == nil {
			rooms = append(rooms, *newRoom)
		}
	}

	return rooms
}
