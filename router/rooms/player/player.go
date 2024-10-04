package player

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	roomId, paramErr := lib.GetIntParam(r, "id")
	if paramErr != nil {
		lib.BadRequest(w, r)
		return
	}
	room := rooms.GetRoom(r.Context(), roomId)

	if room == nil {
		lib.NotFound(w, r)
		return
	}

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
	}

	name := r.PostFormValue("name")
	logger.Info(name)
	lib.NoContent(w, r)
}

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

	logger.Info("Player connected to room %d\n", room.Id())

	eventChan := room.AddPlayer()

	// Listen for client close and delete channel when it happens
	notify := r.Context().Done()
	closeConn := make(chan string)
	go func() {
		<-notify
		fmt.Printf("Player disconnected from room %d\n", room.Id())
		room.RemovePlayer(eventChan)
		closeConn <- "END"
	}()

	// Continuously send data to the client
	go func() {
		for {
			data := <-eventChan
			// This is what's received from a closed channel
			if data == "" {
				break
			}

			logger.Debug("Sending data to player in room %d:\n%s", room.Id(), data)
			fmt.Fprintf(w, data)
			w.(http.Flusher).Flush()
		}
	}()

	// Send initial status
	eventChan <- room.PlayerInitializeEvent()

	// Wait for cleanup to happen and then close the connection
	<-closeConn
}
