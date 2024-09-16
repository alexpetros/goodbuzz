package buzz

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/router/rooms"
	"io"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

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
	room_id, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room := rooms.GetRoom(r.Context(), room_id)

	if room == nil {
		http.NotFound(w, r)
		return
	}

	room.Reset()
	w.WriteHeader(204)
}
