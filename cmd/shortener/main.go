package main

import (
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

var storage = map[string]string{}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", writeDate)
	mux.HandleFunc("/{id}", redirectedHandler)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}

func writeDate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Incorrect Method", 400)
	}

	genId := generateShortUrl()
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read Body", 400)
	}
	storage[genId] = strings.Trim(strings.Trim(string(body), "\n"), "\r")
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte("localhost:8080/" + genId))

}

func redirectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Incorrect Method", 400)
	}
	getUrlId := string(r.URL.Path)[1:]

	data, exist := storage[getUrlId]

	if !exist {
		http.Error(w, "Undefiend ID", 400)
	}

	http.Redirect(w, r, data, http.StatusMovedPermanently)

}

func generateShortUrl() string {
	return uuid.New().String()[:7]
}
