package buzz

import (
	"goodbuzz/lib"
	"goodbuzz/router/rooms"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room := rooms.GetRoom(r.Context(), room_id)
	if room == nil {
		http.NotFound(w, r)
		return
	}

	room.BuzzRoom()
	w.WriteHeader(204)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room := rooms.GetRoom(r.Context(), room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

	room.Reset()
	w.WriteHeader(204)
}
