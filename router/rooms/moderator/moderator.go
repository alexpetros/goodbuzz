package moderator

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms"
	"net/http"
)

func Live(w http.ResponseWriter, r *http.Request) {
	roomId, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}
	room := rooms.GetRoom(r.Context(), roomId)

	if room == nil {
		http.NotFound(w, r)
		return
	}

	// Set the response header to indicate SSE content type
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	logger.Info("Moderator connected to room %d", room.Id())

	eventChan := room.AddModerator()

	// Listen for client close and remove the client from the list
	notify := r.Context().Done()
	closeConn := make(chan string)
	go func() {
		<-notify
		fmt.Printf("Moderator disconnected from room %d", room.Id())
		room.RemoveModerator(eventChan)
		closeConn <- "END"
	}()

	// Continuously send data to the client
	go func() {
		for {
			data := <-eventChan
			if data == "" {
				break
			}

			logger.Debug("Sending data to moderator in room %d:\n%s", room.Id(), data)
			_, err2 := fmt.Fprintf(w, data)
			if err2 != nil {
				logger.Error("Failed to send data to moderatorr in room %d:\n%s", room.Id(), data)
			}
			w.(http.Flusher).Flush()
		}
	}()

	// Send initial status
	room.InitializeModerator(eventChan)

	// Wait for cleanup to happen and then close the connection
	<-closeConn
}
