package rooms

import (
	"goodbuzz/lib"
	"goodbuzz/router/rooms/player"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request) {
	roomId := r.PathValue("id")
	room := lib.GetRoom(roomId)

	if room == nil {
		http.NotFound(w, r)
		return
	}

	if room.Status() == lib.Unlocked {
		res := player.UnlockedBuzzer(room.Id())
		res.Render(r.Context(), w)
	} else {
		w.WriteHeader(204)
	}
}
