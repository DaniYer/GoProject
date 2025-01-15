package app

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

var Ptortage = Links{storage: make(map[string]string)}

func ServiceRun() {
	r := mux.NewRouter()
	r.HandleFunc("/", Ptortage.Shortener).Methods(http.MethodPost)        // POST для сокращения URL
	r.HandleFunc("/{id}", Ptortage.ShortenerLink).Methods(http.MethodGet) // GET для перенаправления

	http.ListenAndServe(":8080", r)

	fmt.Println("Server is running on http://localhost:8080")
}
