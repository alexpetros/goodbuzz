package moderator

import (
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

	// Listen for client close and remove the client from the list
	logger.Info("Moderator connected to room %d", room.Id())
	token, closeConn := room.CreateModerator(w, r)

	// Wait for cleanup to happen and then close the connection
	<-closeConn
	logger.Info("Moderator disconnected from room %d", room.Id())
	room.RemoveModerator(token)
}
