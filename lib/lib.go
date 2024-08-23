package lib

import (
	"fmt"
	"io"
	"net/http"
	"github.com/a-h/templ"
)

func Render(w http.ResponseWriter, r *http.Request, content templ.Component) {
	Base(content).Render(r.Context(), w)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	io.WriteString(w, "<h1>404 Not Found</h1><p><a href=/>Return home</a>")
}

func FormatEvent(event_name string, data string) string {
  return fmt.Sprintf("event: %s\ndata: %s\n\n", event_name, data)
}
