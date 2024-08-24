package router

import (
	"embed"
	"goodbuzz/router/healthcheck"
	"goodbuzz/router/index"
	"goodbuzz/router/rooms/buzz"
	"goodbuzz/router/rooms/moderator"
	"goodbuzz/router/rooms/player"
	"goodbuzz/router/tournaments"
	"net/http"
)

// this bit of go magic embeds everything in the /static directory
//
//go:embed all:static
var content embed.FS

func SetupRouter(mux *http.ServeMux) {

	mux.Handle("/static/", http.FileServer(http.FS(content)))

	mux.HandleFunc("GET /{$}", index.Get)
	mux.HandleFunc("GET /tournaments/{id}", tournaments.Get)

	mux.HandleFunc("GET /rooms/{id}/player", player.Get)
	mux.HandleFunc("GET /rooms/{id}/player/live", player.Live)

	mux.HandleFunc("GET /rooms/{id}/moderator", moderator.Get)
	mux.HandleFunc("GET /rooms/{id}/moderator/live", moderator.Live)

	mux.HandleFunc("PUT /rooms/{id}/buzz", buzz.Put)
	mux.HandleFunc("DELETE /rooms/{id}/buzz", buzz.Delete)

	mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
