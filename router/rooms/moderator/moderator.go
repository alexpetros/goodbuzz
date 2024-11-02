package moderator

import (
	"goodbuzz/lib"
	"goodbuzz/lib/logger"
	"goodbuzz/router/rooms"
	"net/http"

	"github.com/google/uuid"
)

func Live(w http.ResponseWriter, r *http.Request) {
	roomId, err := lib.GetIntParam(r, "id")
	if err != nil {
		lib.BadRequest(w, r)
		return
	}

	room, notFoundErr := rooms.GetRoom(r.Context(), roomId)
	if notFoundErr != nil {
		http.NotFound(w, r)
		return
	}

	token := uuid.NewString()

	// Listen for client close and remove the client from the list
	logger.Info("Moderator %s connected to room %d", token, room.Id)
	room.AttachModerator(w, r, token)

	// Wait for cleanup to happen and then close the connection
	logger.Info("Moderator disconnected from room %d", room.Id)
}
