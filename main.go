package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/limited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, limited\n"))
	})
	mux.HandleFunc("/unlimited", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, unlimited\n"))
	})

	if err := http.ListenAndServe("localhost:8008", mux); err != nil {
		log.Fatalf("Error: %s", err)
	}
}
