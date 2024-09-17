package main

import (
	"errors"
	"fmt"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"goodbuzz/router"
	"net/http"
	"os"
	"os/signal"
)

const DEFAULT_PORT = "3000"
const SQLITE_FILE = "goodbuzz.db"

func main() {
	proxy_port := os.Getenv("BUZZER_PROXY_PORT")
	port := os.Getenv("BUZZER_PORT")

	db.InitDb(SQLITE_FILE)
	mux := http.NewServeMux()
	router.SetupRouter(mux)

	if port == "" {
		port = DEFAULT_PORT
	}

	address := fmt.Sprintf("localhost:%s", port)

	if proxy_port != "" {
		logger.Info("Now listening at http://localhost:%s\n", proxy_port)
	} else {
		logger.Info("Now listening at http://%s\n", address)
	}

	// Start server inside goroutine so that we can listen for an interrupt in the main thread
	go func() {
		err := http.ListenAndServe(address, mux)

		if errors.Is(err, http.ErrServerClosed) {
			logger.Info("Server closed")
		} else if err != nil {
			logger.Fatal("Unexpected error from server: %w", err)
		}
	}()

	// Listen for and handle interrupt signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)

	// Blocks execution until channel receives the signal.
	<-quit

	// Shut the server down and close the database properly
	logger.Info("Shutting server down")
	db.Close()
}
