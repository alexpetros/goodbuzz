package main

import "errors"
import "fmt"
import "io"
import "net/http"
import "os"

const ADDRESS = "localhost:3000"

func main() {
  http.HandleFunc("GET /", healthcheck)

  fmt.Printf("Now listening at http://%s\n", ADDRESS)
  err := http.ListenAndServe(ADDRESS, nil)

  if errors.Is(err, http.ErrServerClosed) {
    fmt.Printf("server closed\n")
  } else if err != nil {
    fmt.Printf("error starting server: %s\n", err)
    os.Exit(1)
  }
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
  	fmt.Printf("Receieved request at /\n")
    io.WriteString(w, "OK")
}
