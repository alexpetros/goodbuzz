package main

import "errors"
import "fmt"
import "net/http"
import "os"

import "goodbuzz/router"

const DEFAULT_PORT = "3000"

func main() {
	proxy_port := os.Getenv("BUZZER_PROXY_PORT")
	port := os.Getenv("BUZZER_PORT")

	mux := http.NewServeMux()
	router.SetupRouter(mux)

  if port == "" {
    port = DEFAULT_PORT
  }

  address := fmt.Sprintf("localhost:%s", port)

	if proxy_port != "" {
		fmt.Printf("Now listening at http://localhost:%s\n", proxy_port)
	} else {
    fmt.Printf("Now listening at http://%s\n", address)
	}

	err := http.ListenAndServe(address, mux)

	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server closed\n")
	} else if err != nil {
		fmt.Printf("error starting server: %s\n", err)
		os.Exit(1)
	}
}
