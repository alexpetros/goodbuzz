package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/a-h/templ"
)

func Render(w http.ResponseWriter, r *http.Request, component templ.Component) {
	Base(component).Render(r.Context(), w)
}

func ToString(component templ.Component) string {
  var buff bytes.Buffer
  component.Render(context.Background(), &buff)
  return buff.String()
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	io.WriteString(w, "<h1>404 Not Found</h1><p><a href=/>Return home</a>")
}

func FormatEvent(event_name string, data string) string {
  return fmt.Sprintf("event: %s\ndata: %s\n\n", event_name, data)
}
