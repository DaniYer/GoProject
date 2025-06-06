package service

import (
	"context"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/repository"
)

type Service struct {
	repo repository.Repository
}

func NewService(repo repository.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveURL(ctx context.Context, userID, originalURL string) (string, error) {
	return s.repo.SaveURL(ctx, userID, originalURL)
}

func (s *Service) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	return s.repo.GetOriginalURL(ctx, shortID)
}

func (s *Service) GetUserURLs(ctx context.Context, userID string) ([]dto.UserURL, error) {
	return s.repo.GetUserURLs(ctx, userID)
}

func (s *Service) SaveBatchURLs(ctx context.Context, userID string, batch []dto.BatchRequestItem, baseURL string) ([]dto.BatchResponseItem, error) {
	return s.repo.SaveBatchURLs(ctx, userID, batch, baseURL)
}
