package main

import (
	"fmt"
	"net/http"

	"github.com/metal-tile/land/firedb"
)

func playerHandler(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")

	ps := firedb.NewPlayerStore()
	v := ps.GetPosition(id)
	fmt.Fprintf(w, "id %s is %+v", id, v)
}
