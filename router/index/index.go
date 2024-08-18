package index

import "net/http"

func Get(w http.ResponseWriter, r *http.Request) {
    ok().Render(r.Context(), w)
}
