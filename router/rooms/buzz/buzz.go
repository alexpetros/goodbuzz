package buzz

import (
	"buzzer/lib"
	"fmt"
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
  res := fmt.Sprintf(
		`<button class="buzzer"
             disabled
             hx-get="/rooms/%d/status"
             hx-trigger="every 500ms"
             >Waiting...
     </button>`,
     room.Id(),
	)

  io.WriteString(w, res)
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
