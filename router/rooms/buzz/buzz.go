package buzz

import (
	"buzzer/lib"
	"io"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room_id := r.PathValue("id")
	room := lib.GetRoom(room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

  room.BuzzRoom()
	io.WriteString(w,
		`<button class="buzzer" disabled>Waiting...</button>`,
	)
}
