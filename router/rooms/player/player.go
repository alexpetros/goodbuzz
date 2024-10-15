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

	room, notFoundErr := rooms.GetRoom(r.Context(), roomId)
	if notFoundErr != nil {
		lib.NotFound(w, r)
		return
	}

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
	}

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}

	name := r.PostFormValue("name")

	room.SetPlayerName(cookie.Value, name)
	lib.NoContent(w, r)
}

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

	// Set the response header to indicate SSE content type
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	logger.Info("Player connected to room %d\n", room.Id())

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}
	token := cookie.Value

	closeConn := room.CreatePlayer(w, r, token)

	// Wait for cleanup to happen and then close the connection
	<-closeConn
	logger.Info("Player disconnected from room %d", room.Id())
	room.RemovePlayer(token)
}
