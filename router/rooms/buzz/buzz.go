package buzz

import (
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/rooms"
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

	token := r.PostFormValue("token")

	room.BuzzRoom(token)
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

	query := r.URL.Query()
	mode := query.Get("mode")

	if mode == "all" {
		room.ResetAll()
	} else if mode == "partial" {
		room.ResetSome()
	} else {
		logger.Warn("bad request with mode %s", mode)
		lib.BadRequest(w, r)
		return
	}

	w.WriteHeader(204)
}
