package main

import "errors"
import "fmt"
import "net/http"
import "os"

import "buzzer/router/index"
import "buzzer/router/healthcheck"

const ADDRESS = "localhost:3000"

func main() {

  proxy_port := os.Getenv("BUZZER_PROXY_PORT")
  mux := http.NewServeMux()

  // TODO static assets
  // https://templ.guide/commands-and-tools/live-reload-with-other-tools#serving-static-assets

  mux.HandleFunc("GET /", index.Get)
  mux.HandleFunc("GET /healthcheck", healthcheck.Get)

  if proxy_port != "" {
    fmt.Printf("Now listening at http://localhost:%s\n", proxy_port)
  } else {
    fmt.Printf("Now listening at http://%s\n", ADDRESS)
  }

  err := http.ListenAndServe(ADDRESS, mux)

  if errors.Is(err, http.ErrServerClosed) {
    fmt.Printf("server closed\n")
  } else if err != nil {
    fmt.Printf("error starting server: %s\n", err)
    os.Exit(1)
  }
}
