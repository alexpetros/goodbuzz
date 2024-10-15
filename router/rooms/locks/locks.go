package locks

import (
	"goodbuzz/lib"
	"goodbuzz/router/rooms"
	"net/http"
)

func Delete(w http.ResponseWriter, r *http.Request) {
	roomId, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := rooms.GetRoom(r.Context(), roomId)
	if notFoundErr != nil {
		http.NotFound(w, r)
		return
	}

	token := r.PathValue("token")
	if token == "" {
		lib.BadRequest(w, r)
		return
	}

	room.UnlockPlayer(token)

	w.WriteHeader(204)
}
