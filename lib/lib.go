package lib

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/a-h/templ"
)

func Render(w http.ResponseWriter, r *http.Request, title string, component templ.Component) {
	Base(title, component).Render(r.Context(), w)
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

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	io.WriteString(w, "<h1>Bad Request</h1><p><a href=/>Return home</a>")
}

func ServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	io.WriteString(w, "<h1>500 Internal Server Error</h1><p><a href=/>Return home</a>")
}

func FormatEvent(event_name string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", event_name, data)
}

func GetIntParam(r *http.Request, param_name string) (id int64, err error) {
	param := r.PathValue(param_name)
	return strconv.ParseInt(param, 10, 64)
}
