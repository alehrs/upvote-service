package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"upvote-service/internal/domain"
	"upvote-service/internal/service"
)

// mockRepo is a manual mock of UpvoteRepository that lets each test
// inject its own behavior via function fields.
type mockRepo struct {
	createFn       func(ctx context.Context, upvote domain.Upvote) (domain.Upvote, error)
	deleteFn       func(ctx context.Context, upvoteID string) error
	getByIdFn      func(ctx context.Context, upvoteID string) (domain.Upvote, error)
	getByArticleFn func(ctx context.Context, articleID string) ([]domain.Upvote, error)
	getByUserFn    func(ctx context.Context, userID string) ([]domain.Upvote, error)
}

func (m *mockRepo) Create(ctx context.Context, upvote domain.Upvote) (domain.Upvote, error) {
	return m.createFn(ctx, upvote)
}
func (m *mockRepo) Delete(ctx context.Context, upvoteID string) error {
	return m.deleteFn(ctx, upvoteID)
}
func (m *mockRepo) GetById(ctx context.Context, upvoteID string) (domain.Upvote, error) {
	return m.getByIdFn(ctx, upvoteID)
}
func (m *mockRepo) GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error) {
	return m.getByArticleFn(ctx, articleID)
}
func (m *mockRepo) GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error) {
	return m.getByUserFn(ctx, userID)
}

func TestCreate_PassesCorrectFieldsToRepository(t *testing.T) {
	var captured domain.Upvote
	repo := &mockRepo{
		createFn: func(_ context.Context, upvote domain.Upvote) (domain.Upvote, error) {
			captured = upvote
			return upvote, nil
		},
	}
	svc := service.New(repo)

	_, err := svc.Create(context.Background(), "article-1", "user-1")
	require.NoError(t, err)

	assert.Equal(t, "article-1", captured.ArticleID)
	assert.Equal(t, "user-1", captured.UserID)
}

func TestCreate_GeneratesValidUUIDv7(t *testing.T) {
	var capturedID string
	repo := &mockRepo{
		createFn: func(_ context.Context, upvote domain.Upvote) (domain.Upvote, error) {
			capturedID = upvote.ID
			return upvote, nil
		},
	}
	svc := service.New(repo)

	_, err := svc.Create(context.Background(), "article-1", "user-1")
	require.NoError(t, err)

	parsed, err := uuid.Parse(capturedID)
	require.NoError(t, err, "ID should be a valid UUID")
	assert.Equal(t, uuid.Version(7), parsed.Version(), "ID should be UUID v7")
}

func TestCreate_EachCallGeneratesUniqueID(t *testing.T) {
	ids := make([]string, 0, 3)
	repo := &mockRepo{
		createFn: func(_ context.Context, upvote domain.Upvote) (domain.Upvote, error) {
			ids = append(ids, upvote.ID)
			return upvote, nil
		},
	}
	svc := service.New(repo)

	for range 3 {
		_, err := svc.Create(context.Background(), "article-1", "user-1")
		require.NoError(t, err)
	}

	assert.Equal(t, 3, len(ids))
	assert.NotEqual(t, ids[0], ids[1])
	assert.NotEqual(t, ids[1], ids[2])
}

func TestCreate_PropagatesDuplicateUpvoteError(t *testing.T) {
	repo := &mockRepo{
		createFn: func(_ context.Context, _ domain.Upvote) (domain.Upvote, error) {
			return domain.Upvote{}, domain.ErrDuplicateUpvote
		},
	}
	svc := service.New(repo)

	_, err := svc.Create(context.Background(), "article-1", "user-1")
	assert.ErrorIs(t, err, domain.ErrDuplicateUpvote)
}

func TestCreate_PropagatesGenericRepositoryError(t *testing.T) {
	repoErr := errors.New("db connection lost")
	repo := &mockRepo{
		createFn: func(_ context.Context, _ domain.Upvote) (domain.Upvote, error) {
			return domain.Upvote{}, repoErr
		},
	}
	svc := service.New(repo)

	_, err := svc.Create(context.Background(), "article-1", "user-1")
	assert.ErrorIs(t, err, repoErr)
}

func TestDelete_DelegatesToRepository(t *testing.T) {
	var capturedID string
	repo := &mockRepo{
		deleteFn: func(_ context.Context, upvoteID string) error {
			capturedID = upvoteID
			return nil
		},
	}
	svc := service.New(repo)

	err := svc.Delete(context.Background(), "upvote-99")
	require.NoError(t, err)
	assert.Equal(t, "upvote-99", capturedID)
}

func TestGetById_ReturnUpvote(t *testing.T) {
	expected := domain.Upvote{ID: "id-1", ArticleID: "a-1", UserID: "u-1", CreatedAt: time.Now()}
	repo := &mockRepo{
		getByIdFn: func(_ context.Context, _ string) (domain.Upvote, error) {
			return expected, nil
		},
	}
	svc := service.New(repo)

	got, err := svc.GetById(context.Background(), "id-1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetByArticle_ReturnsUpvotes(t *testing.T) {
	expected := []domain.Upvote{
		{ID: "id-1", ArticleID: "a-1", UserID: "u-1"},
		{ID: "id-2", ArticleID: "a-1", UserID: "u-2"},
	}
	repo := &mockRepo{
		getByArticleFn: func(_ context.Context, articleID string) ([]domain.Upvote, error) {
			assert.Equal(t, "a-1", articleID)
			return expected, nil
		},
	}
	svc := service.New(repo)

	got, err := svc.GetByArticle(context.Background(), "a-1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}

func TestGetByUser_ReturnsUpvotes(t *testing.T) {
	expected := []domain.Upvote{
		{ID: "id-1", ArticleID: "a-1", UserID: "u-1"},
		{ID: "id-2", ArticleID: "a-2", UserID: "u-1"},
	}
	repo := &mockRepo{
		getByUserFn: func(_ context.Context, userID string) ([]domain.Upvote, error) {
			assert.Equal(t, "u-1", userID)
			return expected, nil
		},
	}
	svc := service.New(repo)

	got, err := svc.GetByUser(context.Background(), "u-1")
	require.NoError(t, err)
	assert.Equal(t, expected, got)
}
