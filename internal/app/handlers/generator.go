package handlers

import (
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/middlewares"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

// NewGenerateShortURLHandler godoc
// @Summary      Создать короткую ссылку
// @Description  Принимает оригинальный URL в теле запроса (в виде текста), возвращает короткую ссылку.
// @Tags         urls
// @Accept       plain
// @Produce      plain
// @Param        url body string true "Оригинальный URL"
// @Success      201 {string} string "Короткая ссылка создана"
// @Failure      400 {string} string "Ошибка чтения тела"
// @Failure      409 {string} string "Ссылка уже существует"
// @Failure      500 {string} string "Ошибка сохранения"
// @Router       / [post]

func NewGenerateShortURLHandler(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		GenerateShortURLHandler(w, r, svc)
	}
}

func GenerateShortURLHandler(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	userID := r.Context().Value(middlewares.UserIDKey).(string)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	originalURL := string(body)

	existingShortURL, err := svc.Store.GetByOriginalURL(originalURL)
	resultURL := svc.BaseURL + "/" + existingShortURL

	if err == nil {
		w.WriteHeader(http.StatusConflict)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(resultURL))
		return
	}

	shortID := service.GenerateRandomID()
	shortID, err = svc.Store.Save(shortID, originalURL, userID)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resultURL = svc.BaseURL + "/" + shortID
	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(resultURL))
}
