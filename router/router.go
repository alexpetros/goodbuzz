package router

import "net/http"
import "embed"

import "buzzer/router/index"
import "buzzer/router/healthcheck"
import "buzzer/router/tournaments"
import "buzzer/router/rooms"
import "buzzer/router/rooms/moderator"

// this bit of go magic embeds everything in the /static directory
//go:embed all:static
var content embed.FS

func SetupRouter (mux *http.ServeMux) {

  mux.Handle("/static/", http.FileServer(http.FS(content)))

  mux.HandleFunc("GET /{$}", index.Get)
  mux.HandleFunc("GET /tournaments/{id}", tournaments.Get)

  mux.HandleFunc("GET /rooms/{id}", rooms.Get)
  mux.HandleFunc("GET /rooms/{id}/moderator", moderator.Get)

  mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
