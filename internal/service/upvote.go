package service

import (
	"context"
	"fmt"
	"upvote-service/internal/domain"

	"github.com/google/uuid"
)

type UpvoteRepository interface {
	Create(ctx context.Context, upvote domain.Upvote) (domain.Upvote, error)
	Delete(ctx context.Context, upvoteID string) error
	GetById(ctx context.Context, upvoteID string) (domain.Upvote, error)
	GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error)
	GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error)
}

type UpvoteService struct {
	repository UpvoteRepository
}

func New(repository UpvoteRepository) *UpvoteService {
	return &UpvoteService{repository: repository}
}

func (s *UpvoteService) Create(ctx context.Context, articleID string, userID string) (domain.Upvote, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return domain.Upvote{}, fmt.Errorf("generate id: %w", err)
	}
	return s.repository.Create(ctx, domain.Upvote{
		ID:        id.String(),
		ArticleID: articleID,
		UserID:    userID,
	})
}

func (s *UpvoteService) Delete(ctx context.Context, upvoteID string) error {
	return s.repository.Delete(ctx, upvoteID)
}

func (s *UpvoteService) GetById(ctx context.Context, upvoteID string) (domain.Upvote, error) {
	return s.repository.GetById(ctx, upvoteID)
}

func (s *UpvoteService) GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error) {
	return s.repository.GetByArticle(ctx, articleID)
}

func (s *UpvoteService) GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error) {
	return s.repository.GetByUser(ctx, userID)
}
