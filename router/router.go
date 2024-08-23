package router

import (
	"goodbuzz/router/healthcheck"
	"goodbuzz/router/index"
	"goodbuzz/router/rooms"
	"goodbuzz/router/rooms/buzz"
	"goodbuzz/router/rooms/moderator"
	"goodbuzz/router/rooms/status"
	"goodbuzz/router/tournaments"
	"embed"
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

	mux.HandleFunc("GET /rooms/{id}", rooms.Get)
	mux.HandleFunc("GET /rooms/{id}/moderator", moderator.Get)
	mux.HandleFunc("GET /rooms/{id}/status", status.Get)
	mux.HandleFunc("PUT /rooms/{id}/buzz", buzz.Put)
	mux.HandleFunc("DELETE /rooms/{id}/buzz", buzz.Delete)

	mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
