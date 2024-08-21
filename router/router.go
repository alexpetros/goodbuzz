package router

import (
	"buzzer/router/healthcheck"
	"buzzer/router/index"
	"buzzer/router/rooms"
	"buzzer/router/rooms/buzz"
	"buzzer/router/rooms/moderator"
	"buzzer/router/tournaments"
	"embed"
	"net/http"
)

// this bit of go magic embeds everything in the /static directory
//go:embed all:static
var content embed.FS

func SetupRouter (mux *http.ServeMux) {

  mux.Handle("/static/", http.FileServer(http.FS(content)))

  mux.HandleFunc("GET /{$}", index.Get)
  mux.HandleFunc("GET /tournaments/{id}", tournaments.Get)

  mux.HandleFunc("GET /rooms/{id}", rooms.Get)
  mux.HandleFunc("GET /rooms/{id}/moderator", moderator.Get)
  mux.HandleFunc("PUT /rooms/{id}/buzz", buzz.Put)

  mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
