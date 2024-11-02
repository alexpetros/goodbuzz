package rooms

import (
	"context"
	"errors"
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
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
	ctx := r.Context()

	room, notFoundErr := GetRoom(ctx, roomId)
	if notFoundErr != nil {
		lib.NotFound(w, r)
		return
	}

	name := r.PostFormValue("name")
	description := r.PostFormValue("description")

	room.SetName(name)
	room.SetDescription(description)
	db.SetRoomNameAndDescription(ctx, roomId, name, description)

	dbRoom := db.GetRoom(ctx, roomId)
	route := fmt.Sprintf("/tournaments/%d/admin", dbRoom.TournamentId)
	lib.HxRedirect(w, r, route)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	roomId, paramErr := lib.GetIntParam(r, "id")
	if paramErr != nil {
		lib.BadRequest(w, r)
		return
	}
	ctx := r.Context()

	room, notFoundErr := GetRoom(ctx, roomId)
	if notFoundErr != nil {
		lib.NotFound(w, r)
		return
	}

	dbRoom := db.GetRoom(ctx, roomId)
	route := fmt.Sprintf("/tournaments/%d/admin", dbRoom.TournamentId)

	db.DeleteRoom(ctx, roomId)
	openRooms.DeleteRoom(roomId)

	room.KickAll()
	lib.HxRedirect(w, r, route)
}

func KickPlayer(w http.ResponseWriter, r *http.Request) {
	roomId, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := GetRoom(r.Context(), roomId)
	if notFoundErr != nil {
		http.NotFound(w, r)
		return
	}

	userToken := r.PathValue("userToken")
	if userToken == "" {
		lib.BadRequest(w, r)
		return
	}

	room.KickPlayer(userToken)

	w.WriteHeader(204)
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
