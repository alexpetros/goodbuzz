package player

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
	}

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}

	userToken := cookie.Value
	name := r.PostFormValue("name")

	room.SetPlayerName(userToken, name)
	db.SetUserName(r.Context(), userToken, name)

	lib.NoContent(w, r)
}

func PutPlayer(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
	}

	userToken := r.PathValue("token")
	name := r.PostFormValue("name")

	room.SetPlayerName(userToken, name)
	db.SetUserName(r.Context(), userToken, name)

	lib.NoContent(w, r)
}

func Live(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}
	userToken := cookie.Value

	if room.IsPlayerAlreadyConnected(userToken) {
		w.Header().Add("Content-Type", "text/event-stream")
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Connection", "keep-alive")
		return
	}

	player := db.GetPlayer(r.Context(), userToken)

	var name string
	if player == nil {
		name = "New Player"
		db.CreatePlayer(r.Context(), userToken, name, room.Id)
	} else {
		name = player.Name
	}

	// If the user has a name cookie that is different than the one in the db, update the name cookie
	// We do this so that users don't revert to their old name when the moderator renames them
	logger.Info("Player %s connected to room %d\n", userToken, room.Id)
	room.AttachPlayer(w, r, userToken, name)

	logger.Info("Player %s disconnected from room %d", userToken, room.Id)
}
