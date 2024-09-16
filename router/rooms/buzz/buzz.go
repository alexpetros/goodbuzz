package buzz

import (
	"fmt"
	"goodbuzz/router/rooms"
	"io"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room_id := r.PathValue("id")
	room := rooms.GetRoom(r.Context(), room_id)

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
	room := rooms.GetRoom(r.Context(), room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

	room.Reset()
	w.WriteHeader(204)
}
