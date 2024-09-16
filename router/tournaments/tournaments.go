package tournaments

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"log"
	"net/http"
)


func Delete(w http.ResponseWriter, r *http.Request) {
  tournament_id, parse_err := lib.GetIntParam(r, "id")

  if parse_err != nil {
    log.Printf("WARN error parsing delete URL %s", parse_err)
    lib.BadRequest(w, r)
    return
  }
  tournament := db.GetTournament(r.Context(), tournament_id)
  if tournament == nil {
    lib.NotFound(w, r)
    return
  }

  delete_err := db.DeleteTournament(r.Context(), tournament_id)
  if delete_err == nil {
    w.Header().Add("HX-Refresh", "true")
  } else {
    lib.ServerError(w, r)
  }
}
