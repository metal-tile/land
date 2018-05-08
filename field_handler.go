package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/metal-tile/land/firedb"
)

func fieldHandler(w http.ResponseWriter, r *http.Request) {
	rowParam := r.FormValue("row")
	colParam := r.FormValue("col")

	row := 0
	if rowParam != "" {
		ro, err := strconv.Atoi(rowParam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "row is %s", err)
		}
		row = ro
	}

	col := 0
	if colParam != "" {
		co, err := strconv.Atoi(colParam)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "col is %s", err)
		}
		row = co
	}

	fs := firedb.NewFieldStore()
	v, err := fs.GetValue(row, col)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "%s", err)
	}
	fmt.Fprintf(w, "%d:%d %+v", row, col, v)
}
