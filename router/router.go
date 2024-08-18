package router

import "net/http"

import "buzzer/router/index"
import "buzzer/router/healthcheck"

func SetupRouter (mux *http.ServeMux) {

  // TODO static assets
  // https://templ.guide/commands-and-tools/live-reload-with-other-tools#serving-static-assets

  mux.HandleFunc("GET /{$}", index.Get)
  mux.HandleFunc("GET /healthcheck", healthcheck.Get)
}
