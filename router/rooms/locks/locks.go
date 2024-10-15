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

	userToken := r.PathValue("userToken")
	if userToken == "" {
		lib.BadRequest(w, r)
		return
	}

	room.UnlockPlayer(userToken)

	w.WriteHeader(204)
}
