package login

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"net/http"
)

func Get(w http.ResponseWriter, r *http.Request) {
	content := login()
	lib.Render(w, r, "Login", content)
}

func Post(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")
	// password := r.PostFormValue("password")

	if noToken != nil {
		cookie = lib.NewUserToken()
	}

	userToken := cookie.Value
	db.LoginMod(r.Context(), userToken)
	http.Redirect(w, r, "/", 303)
}

func Delete(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")

	if noToken != nil {
		cookie = lib.NewUserToken()
	}

	userToken := cookie.Value
	db.DeleteLogin(r.Context(), userToken)
	lib.HxRedirect(w, r, "/")
}
