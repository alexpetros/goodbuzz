package admin

import (
	"context"
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"net/http"
)

func Middleware(next func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenCookie, noToken := r.Cookie("userToken")

		if noToken != nil || !db.IsAdmin(r.Context(), tokenCookie.Value) {
			lib.Forbidden(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), "userToken", tokenCookie.Value)
		r = r.WithContext(ctx)
		next(w, r)
	})
}

func Put(w http.ResponseWriter, r *http.Request) {
	modPassword := r.PostFormValue("mod_password")
	adminPassword := r.PostFormValue("admin_password")
	userToken := r.Context().Value("userToken").(string)

	db.SetSetting(r.Context(), "mod_password", modPassword)
	db.SetSetting(r.Context(), "admin_password", adminPassword)
	db.WipeSessions(r.Context(), userToken)

	lib.HxRedirect(w, r, "/")
}
