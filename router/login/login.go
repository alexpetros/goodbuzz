package login

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"goodbuzz/lib/logger"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request) {
	content := login()
	lib.Render(w, r, "Login", content)
}

func Post(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")
	password := r.PostFormValue("password")
	logger.Info(password)

	if noToken != nil {
		cookie = lib.NewUserToken()
	}

	userToken := cookie.Value
	db.LoginMod(r.Context(), userToken)
	http.Redirect(w, r, "/", 303)
}
