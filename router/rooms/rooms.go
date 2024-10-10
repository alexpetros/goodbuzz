package rooms

import (
	"context"
	"goodbuzz/lib/db"
	"goodbuzz/lib/room"
)

var openRooms = room.NewRoomMap()

func GetRoom(ctx context.Context, roomId int64) *room.Room {
	// TODO handle case where this is nil
	dbRoom := db.GetRoom(ctx, roomId)
	return openRooms.GetOrCreateRoom(dbRoom.Id(), dbRoom.Name())
}

func GetRoomsForTournament(ctx context.Context, tournamentId int64) []room.Room {
	dbRooms := db.GetRoomsForTournament(ctx, tournamentId)
	rooms := make([]room.Room, 0)
	for _, dbRoom := range dbRooms {
		newRoom := GetRoom(ctx, dbRoom.Id())
		rooms = append(rooms, *newRoom)
	}

	return rooms
}
