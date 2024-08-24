package player

import (
	"fmt"
	"goodbuzz/lib"
	"net/http"
)

func Live(w http.ResponseWriter, r *http.Request) {
  param :=  r.PathValue("id")
  room := lib.GetRoom(param)

  if room == nil {
    http.NotFound(w, r)
    return
  }

  // Set the response header to indicate SSE content type
  w.Header().Add("Content-Type", "text/event-stream")
  w.Header().Add("Cache-Control", "no-cache")
  w.Header().Add("Connection", "keep-alive")

  fmt.Printf("Player connected to room %d\n", room.Id())

  eventChan := room.AddPlayer()

  // Delete client when they disconnect
  defer func() {
    room.RemoveListener(eventChan)
    close(eventChan)
  }()

  // Listen for client close and remove the client from the list
  notify := r.Context().Done()
  go func() {
    <-notify
    fmt.Printf("Player disconnected from room %d\n", room.Id())
  }()

  // Continuously send data to the client
  for {
    data := <-eventChan
    fmt.Printf("Sending data to player in room %d\n", room.Id())
    fmt.Fprintf(w, data)
    w.(http.Flusher).Flush()
  }
}
