package player

import (
	"fmt"
	"goodbuzz/lib"
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

	name := r.PostFormValue("name")

	nameCookie := http.Cookie{
		Name:     "name",
		Value:    name,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &nameCookie)

	room.SetPlayerName(cookie.Value, name)
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

	token := r.PathValue("token")
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
	token := cookie.Value

	if room.IsPlayerAlreadyConnected(token) {
		w.Header().Add("Content-Type", "text/event-stream")
		w.Header().Add("Cache-Control", "no-cache")
		w.Header().Add("Connection", "keep-alive")
		fmt.Fprint(w, events.OtherTabOpenEvent(token))
		return
	}

	logger.Info("Player %s connected to room %d\n", token, room.Id())
	room.AttachPlayer(w, r, token)

	logger.Info("Player %s disconnected from room %d", token, room.Id())
}
