package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi"
)

var storage = map[string]string{}

func main() {
	// определяем слайс байт нужной длины
	r := chi.NewRouter()

	r.Post("/", genHandler)
	r.Get("/{id}", shortURL)
	fmt.Println("serve")
	if err := http.ListenAndServe(":8080", r); err != nil {
		panic(err)
	}

}

func genHandler(w http.ResponseWriter, r *http.Request) {

	id := genSym(8)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "NotBody", 307)
		return

	}
	defer r.Body.Close()

	storage[id] = string(body)
	w.WriteHeader(201)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte("http://localhost:8080/" + id))
	fmt.Println(storage)

}
func shortURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	value, exists := storage[id]
	if !exists {
		http.Error(w, "URL not found", 400)
		return
	}
	http.Redirect(w, r, value, http.StatusTemporaryRedirect)
	w.Header().Set("Content-Type", "text/plain")
}

func genSym(len int) string {
	b := make([]byte, len)
	_, err := rand.Read(b) // записываем байты в слайс b
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return ""
	}
	return hex.EncodeToString(b)[:len]

}
