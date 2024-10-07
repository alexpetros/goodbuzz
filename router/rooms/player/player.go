package player

import (
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

	token := r.PostFormValue("token")
	name := r.PostFormValue("name")
	room.SetPlayerName(token, name)
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
	token, closeConn := room.CreatePlayer(w, r)

	// Wait for cleanup to happen and then close the connection
	<-closeConn
	logger.Info("Player disconnected from room %d", room.Id())
	room.RemovePlayer(token)
}
