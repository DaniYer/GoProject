package main

import (
	"io"
	"net/http"

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
		return
	}

	genID := generateShortURL()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Cannot read Body", 400)
	}
	defer r.Body.Close()

	storage[genID] = string(body)
	//strings.Trim(strings.Trim(string(body), "\n"), "\r")
	w.Header().Set("content-type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte("localhost:8080/" + genID))

}

func redirectedHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Incorrect Method", 400)
		return
	}
	getURLID := string(r.URL.Path)[1:]

	data, exist := storage[getURLID]

	if !exist {
		http.Error(w, "Undefiend ID", 400)
	}

	http.Redirect(w, r, data, http.StatusMovedPermanently)

}

func generateShortURL() string {
	return uuid.New().String()[:7]
}
