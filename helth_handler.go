package main

import (
	"fmt"
	"net/http"
)

func helthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, Land")
}
