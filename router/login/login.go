package login

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"net/http"
	"strconv"
)

func Get(w http.ResponseWriter, r *http.Request) {
	content := login()
	lib.Render(w, r, "Login", content)
}

func Post(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")
	password := r.PostFormValue("password")

	if noToken != nil {
		cookie = lib.NewUserToken()
		http.SetCookie(w, cookie)
	}

	userToken := cookie.Value

	if password == db.AdminPassword(r.Context()) {
		db.LoginAdmin(r.Context(), userToken)
		http.Redirect(w, r, "/", 303)
	} else if password == db.ModPassword(r.Context()) {
		db.LoginMod(r.Context(), userToken)
		http.Redirect(w, r, "/", 303)
	} else {
		http.Redirect(w, r, "/login?s=failed", 303)
	}
}

func Delete(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")

	if noToken != nil {
		lib.BadRequest(w, r)
	}

	userToken := cookie.Value
	db.DeleteLogin(r.Context(), userToken)
	lib.HxRedirect(w, r, "/")
}

func PostPlayer(w http.ResponseWriter, r *http.Request) {
	cookie, noToken := r.Cookie("userToken")
	tournament_id_string := r.PostFormValue("tournament_id")
	password := r.PostFormValue("password")

	if noToken != nil {
		cookie = lib.NewUserToken()
		http.SetCookie(w, cookie)
	}

	userToken := cookie.Value

	tournament_id, err := strconv.ParseInt(tournament_id_string, 10, 64)
	if err != nil {
		lib.LoginFailed(w, r)
		return
	}

	tournament := db.GetTournament(r.Context(), tournament_id)
	if tournament == nil {
		lib.LoginFailed(w, r)
		return
	}

	if password == tournament.Password() {
		db.AddUserToTournament(r.Context(), userToken, tournament_id)
		http.Redirect(w, r, tournament.Url(), 303)
	} else {
		lib.LoginFailed(w, r)
	}
}
