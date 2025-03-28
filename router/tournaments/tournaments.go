package tournaments

import (
	"context"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"net/http"
)

func Middleware(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tournament_id, parse_err := lib.GetIntParam(r, "id")
		if parse_err != nil {
			logger.Warn("Error parsing URL %s", parse_err)
			lib.BadRequest(w, r)
			return
		}

		tournament := db.GetTournament(r.Context(), tournament_id)
		if tournament == nil {
			lib.NotFound(w, r)
			return
		}

		isMod := lib.IsMod(r)
		isAdmin := lib.IsAdmin(r)
		isUserAuthed := lib.IsUserAuthed(r, tournament_id)

		if !(isUserAuthed || isMod || isAdmin) {
			lib.Forbidden(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "tournament", tournament)
		ctx = context.WithValue(ctx, "isMod", isMod)
		ctx = context.WithValue(ctx, "isAdmin", isAdmin)
		ctx = context.WithValue(ctx, "isUserAuthed", isUserAuthed)

		r = r.WithContext(ctx)
		next(w, r)
	})
}

func Post(w http.ResponseWriter, r *http.Request) {
	name := r.PostFormValue("name")
	if name == "" {
		lib.BadRequest(w, r)
		return
	}

	err := db.CreateTournament(r.Context(), name)

	if err == nil {
		w.Header().Add("HX-Refresh", "true")
	} else {
		lib.ServerError(w, r)
	}
}

func PostRoom(w http.ResponseWriter, r *http.Request) {
	tournament := r.Context().Value("tournament").(*db.Tournament)
	name := r.PostFormValue("name")
	if name == "" {
		lib.BadRequest(w, r)
		return
	}

	err := db.CreateRoom(r.Context(), tournament.Id(), name)

	if err == nil {
		w.Header().Add("HX-Refresh", "true")
	} else {
		lib.ServerError(w, r)
	}
}

func Put(w http.ResponseWriter, r *http.Request) {
	tournament := r.Context().Value("tournament").(*db.Tournament)
	name := r.PostFormValue("name")
	password := r.PostFormValue("password")

	if name == "" {
		lib.BadRequest(w, r)
		return
	}

	err := db.SetTournamentInfo(r.Context(), tournament.Id(), name, password)

	if err == nil {
		w.Header().Add("HX-Refresh", "true")
	} else {
		lib.ServerError(w, r)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	tournament := r.Context().Value("tournament").(*db.Tournament)
	delete_err := db.DeleteTournament(r.Context(), tournament.Id())
	if delete_err == nil {
		lib.HxRedirect(w, r, "/")
	} else {
		lib.ServerError(w, r)
	}
}
