package repository

import (
	"context"
	"errors"
	"fmt"
	"upvote-service/internal/domain"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UpvoteRepository struct {
	db *pgxpool.Pool
}

func New(db *pgxpool.Pool) *UpvoteRepository {
	return &UpvoteRepository{db: db}
}

func (r *UpvoteRepository) Create(ctx context.Context, upvote domain.Upvote) (domain.Upvote, error) {
	var u domain.Upvote
	err := r.db.QueryRow(ctx,
		`INSERT INTO upvotes (id, article_id, user_id)
		 VALUES ($1, $2, $3)
		 RETURNING id, user_id, article_id, created_at`,
		upvote.ID, upvote.ArticleID, upvote.UserID,
	).Scan(&u.ID, &u.UserID, &u.ArticleID, &u.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return domain.Upvote{}, domain.ErrDuplicateUpvote
		}
		return domain.Upvote{}, fmt.Errorf("create upvote: %w", err)
	}
	return u, nil
}

func (r *UpvoteRepository) Delete(ctx context.Context, upvoteID string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM upvotes WHERE id = $1`, upvoteID)
	if err != nil {
		return fmt.Errorf("delete upvote: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("upvote %s not found", upvoteID)
	}
	return nil
}

func (r *UpvoteRepository) GetById(ctx context.Context, upvoteID string) (domain.Upvote, error) {
	var u domain.Upvote
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, article_id, created_at FROM upvotes WHERE id = $1`,
		upvoteID,
	).Scan(&u.ID, &u.UserID, &u.ArticleID, &u.CreatedAt)
	if err != nil {
		return domain.Upvote{}, fmt.Errorf("get upvote by id: %w", err)
	}
	return u, nil
}

func (r *UpvoteRepository) GetByArticle(ctx context.Context, articleID string) ([]domain.Upvote, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, article_id, created_at FROM upvotes WHERE article_id = $1`,
		articleID,
	)
	if err != nil {
		return nil, fmt.Errorf("get upvotes by article: %w", err)
	}
	defer rows.Close()

	var upvotes []domain.Upvote
	for rows.Next() {
		var u domain.Upvote
		if err := rows.Scan(&u.ID, &u.UserID, &u.ArticleID, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan upvote: %w", err)
		}
		upvotes = append(upvotes, u)
	}
	return upvotes, rows.Err()
}

func (r *UpvoteRepository) GetByUser(ctx context.Context, userID string) ([]domain.Upvote, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, user_id, article_id, created_at FROM upvotes WHERE user_id = $1`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get upvotes by user: %w", err)
	}
	defer rows.Close()

	var upvotes []domain.Upvote
	for rows.Next() {
		var u domain.Upvote
		if err := rows.Scan(&u.ID, &u.UserID, &u.ArticleID, &u.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan upvote: %w", err)
		}
		upvotes = append(upvotes, u)
	}
	return upvotes, rows.Err()
}
