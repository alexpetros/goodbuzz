package status

import (
	"goodbuzz/lib"
	"goodbuzz/router/rooms"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request) {
	room_id := r.PathValue("id")
	room := lib.GetRoom(room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

  if room.Status() == lib.Unlocked {
    res := rooms.UnlockedBuzzer(room.Id())
    res.Render(r.Context(), w)
  }

  w.WriteHeader(204)
  return
}
