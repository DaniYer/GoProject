package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/service"
)

func NewHandleShortenURLv13(svc *service.URLService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		HandleShortenURLv13(w, r, svc)
	}
}

func HandleShortenURLv13(w http.ResponseWriter, r *http.Request, svc *service.URLService) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Ошибка чтения тела", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req dto.ShortenRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Ошибка парсинга JSON", http.StatusBadRequest)
		return
	}

	shortID, isConflict, err := svc.ShortenJSON(req.URL)
	if err != nil {
		http.Error(w, "Ошибка сохранения", http.StatusInternalServerError)
		return
	}

	resp := dto.ShortenResponse{Result: svc.BaseURL + "/" + shortID}

	w.Header().Set("Content-Type", "application/json")

	if isConflict {
		w.WriteHeader(http.StatusConflict)
	} else {
		w.WriteHeader(http.StatusCreated)
	}

	json.NewEncoder(w).Encode(resp)
}
