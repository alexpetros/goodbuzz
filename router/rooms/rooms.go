package rooms

import (
	"context"
	"errors"
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room"
	"net/http"
)

var openRooms = room.NewRoomMap()

func Put(w http.ResponseWriter, r *http.Request) {
	roomId, paramErr := lib.GetIntParam(r, "id")
	if paramErr != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := GetRoom(r.Context(), roomId)
	if notFoundErr != nil {
		lib.NotFound(w, r)
		return
	}

	description := r.PostFormValue("description")
	logger.Info(description)
	room.SetDescription(description)
	db.SetRoomDescription(r.Context(), roomId, description)

	route := fmt.Sprintf("/rooms/%d/moderator", roomId)
	lib.HxRedirect(w, r, route)
}

func GetRoom(ctx context.Context, roomId int64) (*room.Room, error) {
	dbRoom := db.GetRoom(ctx, roomId)
	if dbRoom == nil {
		return nil, errors.New("room not found")
	}

	return openRooms.GetOrCreateRoom(dbRoom.RoomId, dbRoom.Name, dbRoom.Description), nil
}

func GetRoomsForTournament(ctx context.Context, tournamentId int64) []room.Room {
	dbRooms := db.GetRoomsForTournament(ctx, tournamentId)
	rooms := make([]room.Room, 0)
	for _, dbRoom := range dbRooms {
		newRoom, notFoundErr := GetRoom(ctx, dbRoom.RoomId)
		if notFoundErr == nil {
			rooms = append(rooms, *newRoom)
		}
	}

	return rooms
}
