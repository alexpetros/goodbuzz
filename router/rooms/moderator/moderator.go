package moderator

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms"
	"net/http"
)

func Live(w http.ResponseWriter, r *http.Request) {
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

	// Set the response header to indicate SSE content type
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	logger.Info("Moderator connected to room %d\n", room.Id())

	eventChan := room.AddModerator()

	// Listen for client close and remove the client from the list
	notify := r.Context().Done()
	go func() {
		<-notify
		fmt.Printf("Moderator disconnected from room %d\n", room.Id())
		room.RemoveModerator(eventChan)
		close(eventChan)
	}()

	// Continuously send data to the client
	for {
		data := <-eventChan
		if data == "" {
			break
		}

		logger.Debug("Sending data to moderator in room %d:\n%s", room.Id(), data)
		fmt.Fprintf(w, data)
		w.(http.Flusher).Flush()
	}
}
