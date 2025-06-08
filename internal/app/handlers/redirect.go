package handlers

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// URLStore описывает методы для сохранения и получения URL.
type URLStore interface {
	Save(shortURL, originalURL string) (string, error)
	Get(shortURL string) (string, error)
}

func NewRedirectToOriginalURL(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		RedirectToOriginalURL(w, r, svc)
	}
}

func RedirectToOriginalURL(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	id := r.PathValue("id")

	originalURL, err := svc.Store.Get(id)
	if err != nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)

}
