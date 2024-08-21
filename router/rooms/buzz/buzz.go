package buzz

import (
	"buzzer/router/rooms"
	"io"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
    room_id :=  r.PathValue("id")
    room := rooms.GetRoom(room_id)

    if room == nil {
      http.NotFound(w, r)
      return
    }

    io.WriteString(w,
    `<button class="buzzer" disabled>Waiting...</button>`,
  )
}
