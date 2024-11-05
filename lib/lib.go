package lib

import (
	"bytes"
	"context"
	"fmt"
	"goodbuzz/lib/db"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

func Render(w http.ResponseWriter, r *http.Request, title string, component templ.Component) {
	Base(title, component).Render(r.Context(), w)
}

func ToString(component templ.Component) string {
	var buff bytes.Buffer
	component.Render(context.Background(), &buff)
	return buff.String()
}

func NoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(204)
}

func NotFound(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	io.WriteString(w, "<h1>404 Not Found</h1><p><a href=/>Return home</a>")
}

func HxRedirect(w http.ResponseWriter, r *http.Request, route string) {
	w.Header().Add("Hx-Redirect", route)
	w.WriteHeader(200)
}

func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(400)
	io.WriteString(w, "<h1>Bad Request</h1><p><a href=/>Return home</a>")
}

func ServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(500)
	io.WriteString(w, "<h1>500 Internal Server Error</h1><p><a href=/>Return home</a>")
}

func GetIntParam(r *http.Request, param_name string) (id int64, err error) {
	param := r.PathValue(param_name)
	return strconv.ParseInt(param, 10, 64)
}

func FormatEventString(eventName string, data string) string {
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func FormatEventComponent(eventName string, component templ.Component) string {
	data := ToString(component)
	return fmt.Sprintf("event: %s\ndata: %s\n\n", eventName, data)
}

func CombineEvents(events ...string) string {
	var sb strings.Builder
	for _, message := range events {
		sb.WriteString(message)
	}

	return sb.String()
}

func NewUserToken() *http.Cookie {
	return &http.Cookie {
		Name:     "userToken",
		Value:    uuid.NewString(),
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func IsMod(r *http.Request) bool {
	isMod := false

	cookie, noToken := r.Cookie("userToken")
	if noToken == nil {
		isMod = db.IsMod(r.Context(), cookie.Value)
	}

	return isMod
}

func IsAdmin(r *http.Request) bool {
	isAdmin := false

	cookie, noToken := r.Cookie("userToken")
	if noToken == nil {
		isAdmin = db.IsAdmin(r.Context(), cookie.Value)
	}

	return isAdmin
}
