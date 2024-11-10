package admin

import (
	"goodbuzz/lib"
	"goodbuzz/lib/db"
	"net/http"
)

func Put(w http.ResponseWriter, r *http.Request) {
	modPassword := r.PostFormValue("mod_password")
	adminPassword := r.PostFormValue("admin_password")

	db.SetSetting(r.Context(), "mod_password", modPassword)
	db.SetSetting(r.Context(), "admin_password", adminPassword)

	lib.HxRedirect(w, r, "/")
}
