package service

import (
	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

type URLService struct {
	Store   URLStore
	BaseURL string
}

type URLStore interface {
	Save(shortURL, originalURL, userID string) (string, error)
	Get(shortURL string) (string, error)
	GetByOriginalURL(originalURL string) (string, error)
	GetAllByUser(userID string) ([]dto.UserURL, error)
	BatchDelete(userID string, shortURLs []string) error
}

// Plain text shorten (POST "/")
func (s *URLService) Shorten(originalURL, userID string) (string, bool, error) {
	existingShortURL, err := s.Store.GetByOriginalURL(originalURL)
	if err == nil {
		return existingShortURL, true, nil
	}

	shortID := GenerateRandomID()
	shortID, err = s.Store.Save(shortID, originalURL, userID)
	if err != nil {
		return "", false, err
	}
	return shortID, false, nil
}

// JSON shorten (POST "/api/shorten")
func (s *URLService) ShortenJSON(originalURL, userID string) (string, bool, error) {
	existingShortURL, err := s.Store.GetByOriginalURL(originalURL)
	if err == nil {
		return existingShortURL, true, nil
	}

	shortID := GenerateRandomID()
	shortID, err = s.Store.Save(shortID, originalURL, userID)
	if err != nil {
		return "", false, err
	}
	return shortID, false, nil
}

// Batch shorten (POST "/api/shorten/batch")
func (s *URLService) ShortenBatch(requests []dto.BatchRequest, userID string) ([]dto.BatchResponse, error) {
	responses := make([]dto.BatchResponse, len(requests))

	for i, req := range requests {
		shortURL := GenerateRandomID()
		if _, err := s.Store.Save(shortURL, req.OriginalURL, userID); err != nil {
			return nil, err
		}
		responses[i] = dto.BatchResponse{
			CorrelationID: req.CorrelationID,
			ShortURL:      s.BaseURL + "/" + shortURL,
		}
	}
	return responses, nil
}

// User URLs (GET "/api/user/urls")
func (s *URLService) GetAllUserURLs(userID string) ([]dto.UserURL, error) {
	urls, err := s.Store.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}
	for i := range urls {
		urls[i].ShortURL = s.BaseURL + "/" + urls[i].ShortURL
	}
	return urls, nil
}

func (s *URLService) BatchDelete(userID string, shortURLs []string) error {
	return s.Store.BatchDelete(userID, shortURLs)
}
