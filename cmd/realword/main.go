package main

import (
	"log"
	"net/http"
	"realworld/internal/app"
)

func main() {
	handler := app.GetApp()
	listenAddr := ":8080"
	log.Printf("starting listening server at %s", listenAddr)
	http.ListenAndServe("/", handler)
}
