package handlers

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/service"
	"github.com/go-chi/chi/v5"
)

// NewRedirectToOriginalURL godoc
// @Summary      Перенаправление по короткой ссылке
// @Description  Получает оригинальный URL по его короткому идентификатору и делает перенаправление.
// @Tags         redirect
// @Param        id   path      string  true  "Короткий идентификатор ссылки"
// @Success      307  {string}  string  "Temporary Redirect"
// @Failure      404  {string}  string  "URL not found"
// @Failure      410  {string}  string  "URL deleted"
// @Router       /{id} [get]
func NewRedirectToOriginalURL(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortURL := chi.URLParam(r, "id")

		originalURL, err := svc.Store.Get(shortURL)
		if err != nil {
			if err.Error() == "gone" {
				http.Error(w, "URL deleted", http.StatusGone)
				return
			}
			http.Error(w, "URL not found", http.StatusNotFound)
			return
		}

		http.Redirect(w, r, originalURL, http.StatusTemporaryRedirect)
	}
}
