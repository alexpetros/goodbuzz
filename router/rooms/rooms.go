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

func Middleware(next func (http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		ctx := context.WithValue(r.Context(), "room", room)
		r = r.WithContext(ctx)
		next(w, r)
	})
}


func Put(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	room := ctx.Value("room").(*room.Room)
	name := r.PostFormValue("name")
	description := r.PostFormValue("description")

	room.SetName(name)
	room.SetDescription(description)
	db.SetRoomNameAndDescription(ctx, room.Id, name, description)

	dbRoom := db.GetRoom(ctx, room.Id)
	route := fmt.Sprintf("/tournaments/%d/admin", dbRoom.TournamentId)
	lib.HxRedirect(w, r, route)
}

func Description(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	room := ctx.Value("room").(*room.Room)
	description := r.PostFormValue("description")

	room.SetDescription(description)
	db.SetRoomNameAndDescription(ctx, room.Id, room.Name, description)

	response := "<button hx-trigger=\"load delay:1s\" hx-on::trigger=\"this.innerText='Save'\">Saved!</button>"
	fmt.Fprintf(w, response)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	room := ctx.Value("room").(*room.Room)
	dbRoom := db.GetRoom(ctx, room.Id)
	route := fmt.Sprintf("/tournaments/%d/admin", dbRoom.TournamentId)

	db.DeleteRoom(ctx, room.Id)
	openRooms.DeleteRoom(room.Id)

	room.KickAll()
	lib.HxRedirect(w, r, route)
}

func KickPlayer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	room := ctx.Value("room").(*room.Room)

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
