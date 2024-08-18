package index

import "fmt"
import "io"
import "net/http"

func Get(w http.ResponseWriter, r *http.Request) {
  	fmt.Printf("Receieved request at /\n")
    io.WriteString(w, "OK")
}
