package player

import (
	"fmt"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/events"
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

	userToken := cookie.Value
	name := r.PostFormValue("name")

	http.SetCookie(w, makeNameCookie(name))
	room.SetPlayerName(userToken, name)
	db.SetUserName(r.Context(), userToken, name)

	lib.NoContent(w, r)
}

func PutPlayer(w http.ResponseWriter, r *http.Request) {
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

	userToken := r.PathValue("token")
	name := r.PostFormValue("name")

	room.SetPlayerName(userToken, name)
	db.SetUserName(r.Context(), userToken, name)

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

	nameCookie, err := r.Cookie("name")

	var name string
	if err != nil {
		name = "New Player"
	} else {
		name = nameCookie.Value
	}

	// If the user has a name cookie that is different than the one in the db, update the name cookie
	// We do this so that users don't revert to their old name when the moderator renames them
	dbName := db.GetName(r.Context(), userToken)
	if dbName != "" && dbName != name {
		name = dbName
		// Update the player's name cookie to the new name
		http.SetCookie(w, makeNameCookie(name))
	}


	logger.Info("Player %s connected to room %d\n", userToken, room.Id)
	room.AttachPlayer(w, r, userToken, name)

	logger.Info("Player %s disconnected from room %d", userToken, room.Id)
}

func makeNameCookie(name string) *http.Cookie {
	return &http.Cookie{
		Name:     "name",
		Value:    name,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}
