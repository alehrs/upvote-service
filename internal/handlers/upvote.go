package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"upvote-service/internal/domain"

	"github.com/go-chi/chi/v5"
)

type UpvoteService interface {
	Create(ctx context.Context, articleID string, userID string) (domain.Upvote, error)
	Delete(ctx context.Context, upvoteID string) error
	GetById(ctx context.Context, upvoteID string) (domain.Upvote, error)
	GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error)
	GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error)
}

type UpvoteHandler struct {
	service UpvoteService
}

func New(service UpvoteService) *UpvoteHandler {
	return &UpvoteHandler{service: service}
}

func (h *UpvoteHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ArticleID string `json:"article_id"`
		UserID    string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondErr(w, http.StatusBadRequest, "invalid request body")
		return
	}

	upvote, err := h.service.Create(r.Context(), body.ArticleID, body.UserID)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicateUpvote) {
			respondErr(w, http.StatusConflict, err.Error())
			return
		}
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusCreated, upvote)
}

func (h *UpvoteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	upvoteID := chi.URLParam(r, "upvoteID")

	if err := h.service.Delete(r.Context(), upvoteID); err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// List supports ?article_id=... or ?user_id=... query params.
func (h *UpvoteHandler) List(w http.ResponseWriter, r *http.Request) {
	articleID := r.URL.Query().Get("article_id")
	userID := r.URL.Query().Get("user_id")

	var (
		upvotes []domain.Upvote
		err     error
	)

	switch {
	case articleID != "":
		upvotes, err = h.service.GetByArticle(r.Context(), articleID)
	case userID != "":
		upvotes, err = h.service.GetByUser(r.Context(), userID)
	default:
		respondErr(w, http.StatusBadRequest, "article_id or user_id query param required")
		return
	}

	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, upvotes)
}

func (h *UpvoteHandler) Get(w http.ResponseWriter, r *http.Request) {
	upvoteID := chi.URLParam(r, "upvoteID")

	upvote, err := h.service.GetById(r.Context(), upvoteID)
	if err != nil {
		respondErr(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, upvote)
}

func respondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func respondErr(w http.ResponseWriter, status int, msg string) {
	respondJSON(w, status, map[string]string{"error": msg})
}
