package main

import (
	"fmt"
	"net/http"
)

func fieldHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, World")
}
