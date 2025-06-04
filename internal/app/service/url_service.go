package service

import (
	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/randomid"
)

// Структура сервиса
type URLService struct {
	Store   URLStore
	BaseURL string
}
type URLStore interface {
	Save(shortURL, originalURL string) (string, error)
	Get(shortURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
}

// Бизнес-логика для батча
func (s *URLService) ShortenBatch(requests []dto.BatchRequest) ([]dto.BatchResponse, error) {
	responses := make([]dto.BatchResponse, len(requests))

	for i, req := range requests {
		shortURL := randomid.GenerateRandomID()
		if _, err := s.Store.Save(shortURL, req.OriginalURL); err != nil {
			return nil, err
		}
		responses[i] = dto.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      s.BaseURL + "/" + shortURL,
		}
	}
	return responses, nil
}

// Single shorten service (для plain text запросов)
func (s *URLService) Shorten(originalURL string) (string, bool, error) {
	shortID := randomid.GenerateRandomID()
	shortID, err := s.Store.Save(shortID, originalURL)
	if err != nil {
		return "", false, err
	}

	// Если короткий URL уже существовал — мы определяем это так:
	existingShortURL, _ := s.Store.GetByOriginalURL(originalURL)
	if existingShortURL == shortID {
		return shortID, false, nil
	}
	return shortID, true, nil
}

// JSON shorten service
func (s *URLService) ShortenJSON(originalURL string) (string, bool, error) {
	existingShortURL, err := s.Store.GetByOriginalURL(originalURL)
	if err == nil {
		return existingShortURL, true, nil
	}

	shortID := randomid.GenerateRandomID()

	shortID, err = s.Store.Save(shortID, originalURL)
	if err != nil {
		return "", false, err
	}
	return shortID, false, nil

}
