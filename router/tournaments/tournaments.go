package tournaments

import (
	"context"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"net/http"
)

func Middleware(next func (http.ResponseWriter, *http.Request)) http.Handler {
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

		ctx := context.WithValue(r.Context(), "tournament", tournament)
		r = r.WithContext(ctx)
		next(w, r)
	})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	tournament := r.Context().Value("tournament").(*db.Tournament)
	delete_err := db.DeleteTournament(r.Context(), tournament.Id())
	if delete_err == nil {
		w.Header().Add("HX-Refresh", "true")
	} else {
		lib.ServerError(w, r)
	}
}
