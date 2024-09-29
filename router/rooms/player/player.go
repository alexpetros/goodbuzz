package player

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

	logger.Info("Player connected to room %d\n", room.Id())

	eventChan := room.AddPlayer()

	// Listen for client close and delete channel when it happens
	notify := r.Context().Done()
	go func() {
		<-notify
		fmt.Printf("Player disconnected from room %d\n", room.Id())
		room.RemovePlayer(eventChan)
	}()

	// Send initial status
	event := room.GetCurrentStatus()
	logger.Debug("Sending data to player in room %d:\n%s", room.Id(), event)
	fmt.Fprintf(w, event)
	w.(http.Flusher).Flush()

	// Continuously send data to the client
	for {
		data := <-eventChan
		// This is what's receieved from a closed channel
		if data == "" {
			break
		}

		logger.Debug("Sending data to player in room %d:\n%s", room.Id(), data)
		fmt.Fprintf(w, data)
		w.(http.Flusher).Flush()
	}
}
