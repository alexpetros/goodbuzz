package buzz

import (
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room"
	"goodbuzz/router/rooms"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}

	resetToken := r.PostFormValue("resetToken")

	room.BuzzRoom(cookie.Value, resetToken)
	w.WriteHeader(204)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

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
