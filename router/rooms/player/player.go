package player

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/events"
	"goodbuzz/lib/logger"
	"goodbuzz/lib/room"
	"net/http"
	"strconv"
)

func Put(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
		return
	}

	cookie, noToken := r.Cookie("userToken")
	if noToken != nil {
		lib.BadRequest(w, r)
		return
	}

	userToken := cookie.Value
	name := r.PostFormValue("name")
	if name == "" {
		name = "New Player"
	}

	team, notInt := strconv.ParseInt(r.PostFormValue("team"), 10, 64)
	if notInt != nil {
		lib.BadRequest(w, r)
		return
	}

	room.UpdatePlayer(userToken, name, team)
	db.UpdatePlayer(r.Context(), userToken, name, team)

	lib.NoContent(w, r)
}

func PutPlayer(w http.ResponseWriter, r *http.Request) {
	room := r.Context().Value("room").(*room.Room)

	formErr := r.ParseForm()
	if formErr != nil {
		lib.BadRequest(w, r)
		return
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
		fmt.Fprint(w, events.OtherTabOpenEvent(userToken))
		return
	}

	player := db.GetPlayer(r.Context(), userToken)

	var name string
	var team int64
	if player == nil {
		name = "New Player"
		team = 1
		db.CreatePlayer(r.Context(), userToken, name, team, room.Id)
	} else {
		name = player.Name
		team = player.Team
	}

	// If the user has a name cookie that is different than the one in the db, update the name cookie
	// We do this so that users don't revert to their old name when the moderator renames them
	logger.Info("Player %s connected to room %d\n", userToken, room.Id)
	room.AttachPlayer(w, r, userToken, name, team)

	logger.Info("Player %s disconnected from room %d", userToken, room.Id)
}
