package locks

import (
	"goodbuzz/lib"
	"goodbuzz/lib/room"
	"net/http"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	userToken := r.PathValue("userToken")
	if userToken == "" {
		lib.BadRequest(w, r)
		return
	}

	room.UnlockPlayer(userToken)

	w.WriteHeader(204)
}
