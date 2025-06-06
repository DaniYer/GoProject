package repository

import (
	"context"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
)

// Repository defines the interface for URL storage operations.
type Repository interface {
	SaveURL(ctx context.Context, userID, originalURL string) (string, error)
	GetOriginalURL(ctx context.Context, shortID string) (string, error)
	GetUserURLs(ctx context.Context, userID string) ([]dto.UserURL, error)
	SaveBatchURLs(ctx context.Context, userID string, batch []dto.BatchRequestItem, baseURL string) ([]dto.BatchResponseItem, error)
}
