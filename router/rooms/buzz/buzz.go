package buzz

import (
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := rooms.GetRoom(r.Context(), room_id)
	if notFoundErr != nil {
		http.NotFound(w, r)
		return
	}

	token, noToken := r.Cookie("token")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}

	resetToken := r.PostFormValue("resetToken")

	room.BuzzRoom(token.Value, resetToken)
	w.WriteHeader(204)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := rooms.GetRoom(r.Context(), room_id)
	if notFoundErr != nil {
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
