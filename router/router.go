package router

import "net/http"
import "embed"

import "buzzer/router/index"
import "buzzer/router/healthcheck"

// this bit of go magic embeds all the static files
//go:embed all:static
var content embed.FS

func SetupRouter (mux *http.ServeMux) {

  mux.Handle("/static/", http.FileServer(http.FS(content)))

  mux.HandleFunc("GET /{$}", index.Get)
  mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
