package redirect

import (
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/storage"
	"github.com/go-chi/chi/v5"
)

func RedirectToOriginalURL(w http.ResponseWriter, r *http.Request, read *storage.Consumer) {
	data, _ := read.ReadEvents()

	// получаем короткий идентификатор из URL
	shortID := chi.URLParam(r, "id")

	for _, value := range data {
		if value.ShortURL == shortID {
			http.Redirect(w, r, value.OriginalURL, http.StatusTemporaryRedirect)
			return // Завершаем выполнение, если найдено соответствие
		}
	}

	http.Error(w, "URL not found", http.StatusBadRequest)
}
