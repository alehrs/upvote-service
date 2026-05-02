package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"upvote-service/internal/domain"
	"upvote-service/internal/handlers"
)

type mockService struct {
	createFn       func(ctx context.Context, articleID string, userID string) (domain.Upvote, error)
	deleteFn       func(ctx context.Context, upvoteID string) error
	getByIdFn      func(ctx context.Context, upvoteID string) (domain.Upvote, error)
	getByArticleFn func(ctx context.Context, articleID string) ([]domain.Upvote, error)
	getByUserFn    func(ctx context.Context, userID string) ([]domain.Upvote, error)
}

func (m *mockService) Create(ctx context.Context, articleID string, userID string) (domain.Upvote, error) {
	return m.createFn(ctx, articleID, userID)
}
func (m *mockService) Delete(ctx context.Context, upvoteID string) error {
	return m.deleteFn(ctx, upvoteID)
}
func (m *mockService) GetById(ctx context.Context, upvoteID string) (domain.Upvote, error) {
	return m.getByIdFn(ctx, upvoteID)
}
func (m *mockService) GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error) {
	return m.getByArticleFn(ctx, articleID)
}
func (m *mockService) GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error) {
	return m.getByUserFn(ctx, userID)
}

// routerWith registers the handler on a chi router so URL params are resolved correctly.
func routerWith(h *handlers.UpvoteHandler) http.Handler {
	r := chi.NewRouter()
	r.Post("/upvotes", h.Create)
	r.Delete("/upvotes/{upvoteID}", h.Delete)
	r.Get("/upvotes", h.List)
	r.Get("/upvotes/{upvoteID}", h.Get)
	return r
}

func TestCreate_Success(t *testing.T) {
	expected := domain.Upvote{ID: "uuid-1", ArticleID: "a-1", UserID: "u-1", CreatedAt: time.Now()}
	svc := &mockService{
		createFn: func(_ context.Context, articleID, userID string) (domain.Upvote, error) {
			assert.Equal(t, "a-1", articleID)
			assert.Equal(t, "u-1", userID)
			return expected, nil
		},
	}
	r := routerWith(handlers.New(svc))

	body := `{"article_id":"a-1","user_id":"u-1"}`
	req := httptest.NewRequest(http.MethodPost, "/upvotes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var got domain.Upvote
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, expected.ID, got.ID)
}

func TestCreate_InvalidJSON(t *testing.T) {
	r := routerWith(handlers.New(&mockService{}))

	req := httptest.NewRequest(http.MethodPost, "/upvotes", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreate_DuplicateUpvote(t *testing.T) {
	svc := &mockService{
		createFn: func(_ context.Context, _, _ string) (domain.Upvote, error) {
			return domain.Upvote{}, domain.ErrDuplicateUpvote
		},
	}
	r := routerWith(handlers.New(svc))

	body := `{"article_id":"a-1","user_id":"u-1"}`
	req := httptest.NewRequest(http.MethodPost, "/upvotes", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestDelete_Success(t *testing.T) {
	var capturedID string
	svc := &mockService{
		deleteFn: func(_ context.Context, upvoteID string) error {
			capturedID = upvoteID
			return nil
		},
	}
	r := routerWith(handlers.New(svc))

	req := httptest.NewRequest(http.MethodDelete, "/upvotes/uuid-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "uuid-1", capturedID)
}

func TestGet_Success(t *testing.T) {
	expected := domain.Upvote{ID: "uuid-1", ArticleID: "a-1", UserID: "u-1"}
	svc := &mockService{
		getByIdFn: func(_ context.Context, upvoteID string) (domain.Upvote, error) {
			assert.Equal(t, "uuid-1", upvoteID)
			return expected, nil
		},
	}
	r := routerWith(handlers.New(svc))

	req := httptest.NewRequest(http.MethodGet, "/upvotes/uuid-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got domain.Upvote
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Equal(t, expected.ID, got.ID)
}

func TestList_ByArticle(t *testing.T) {
	expected := []domain.Upvote{{ID: "uuid-1", ArticleID: "a-1", UserID: "u-1"}}
	svc := &mockService{
		getByArticleFn: func(_ context.Context, articleID string) ([]domain.Upvote, error) {
			assert.Equal(t, "a-1", articleID)
			return expected, nil
		},
	}
	r := routerWith(handlers.New(svc))

	req := httptest.NewRequest(http.MethodGet, "/upvotes?article_id=a-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var got []domain.Upvote
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	assert.Len(t, got, 1)
}

func TestList_ByUser(t *testing.T) {
	expected := []domain.Upvote{{ID: "uuid-1", ArticleID: "a-1", UserID: "u-1"}}
	svc := &mockService{
		getByUserFn: func(_ context.Context, userID string) ([]domain.Upvote, error) {
			assert.Equal(t, "u-1", userID)
			return expected, nil
		},
	}
	r := routerWith(handlers.New(svc))

	req := httptest.NewRequest(http.MethodGet, "/upvotes?user_id=u-1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestList_MissingQueryParams(t *testing.T) {
	r := routerWith(handlers.New(&mockService{}))

	req := httptest.NewRequest(http.MethodGet, "/upvotes", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
