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

func Delete(w http.ResponseWriter, r *http.Request) {
	room_id := r.PathValue("id")
	room := lib.GetRoom(room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

  room.Reset()

  w.Header().Add("HX-Retarget", ".buzzer-status")
  io.WriteString(w, `<span class="buzzer-status">Unlocked</span>`)
}
