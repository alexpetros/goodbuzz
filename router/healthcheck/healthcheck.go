package healthcheck

import "io"
import "net/http"

func Get(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "OK")
}
