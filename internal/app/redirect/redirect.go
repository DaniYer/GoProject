package redirect

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// URLStore описывает методы для сохранения и получения URL.
type URLStore interface {
	Save(shortURL, originalURL string) (string, error)
	Get(shortURL string) (string, error)
}

func NewRedirectToOriginalURL(read URLStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RedirectToOriginalURL(w, r, read)
	}
}
func RedirectToOriginalURL(w http.ResponseWriter, r *http.Request, read URLStore) {
	if r.Method != http.MethodGet {
		http.Error(w, "Unresolved method", http.StatusBadRequest)
		return
	}
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "Empty path or not found", http.StatusBadRequest)
		return
	}
	originalURL, err := read.Get(id)
	if err != nil || originalURL == "" {
		http.Error(w, "URL not found", http.StatusBadRequest)
		return
	}
	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)

}
