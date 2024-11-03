package router

import (
	"embed"
	"goodbuzz/router/admin"
	"goodbuzz/router/healthcheck"
	"goodbuzz/router/index"
	"goodbuzz/router/login"
	"goodbuzz/router/rooms"
	"goodbuzz/router/rooms/buzz"
	"goodbuzz/router/rooms/locks"
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
	mux.Handle(			"/static/", http.FileServer(http.FS(content)))
	mux.HandleFunc( "GET /{$}", index.Get)
	mux.HandleFunc(	"GET /login", login.Get)
	mux.HandleFunc(	"POST /login", login.Post)
	mux.HandleFunc(	"DELETE /login", login.Delete)

	mux.Handle(			"GET /tournaments/{id}", 						tournaments.Middleware(tournaments.Get))
	mux.Handle(			"DELETE /tournaments/{id}", 				tournaments.Middleware(tournaments.Delete))

	mux.Handle(			"PUT /rooms/{id}", 									rooms.Middleware(rooms.Put))
	mux.Handle(			"DELETE /rooms/{id}", 							rooms.Middleware(rooms.Delete))
	mux.Handle(			"GET /rooms/{id}/edit", 						rooms.Middleware(rooms.Get))

	mux.Handle(			"GET /rooms/{id}/player", 					rooms.Middleware(player.Get))
	mux.Handle(			"GET /rooms/{id}/player/live", 			rooms.Middleware(player.Live))
	mux.Handle(			"PUT /rooms/{id}/player", 					rooms.Middleware(player.Put))
	mux.Handle(			"PUT /rooms/{id}/player/{token}", 	rooms.Middleware(player.PutPlayer))

	mux.Handle(			"GET /rooms/{id}/moderator", 				rooms.Middleware(moderator.Get))
	mux.Handle(			"GET /rooms/{id}/moderator/live", 	rooms.Middleware(moderator.Live))

	mux.HandleFunc(	"PUT /rooms/{id}/buzz", 					buzz.Put)

	mux.HandleFunc(	"DELETE /rooms/{id}/buzz", 				buzz.Delete)
	mux.HandleFunc(	"DELETE /rooms/{id}/players/{userToken}", rooms.KickPlayer)
	mux.HandleFunc(	"DELETE /rooms/{id}/locks/{userToken}", locks.Delete)

	mux.HandleFunc(	"GET /admin", admin.Get)

	mux.HandleFunc(	"GET /healthcheck", healthcheck.Get)
}
