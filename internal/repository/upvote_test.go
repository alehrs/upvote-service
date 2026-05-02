package repository_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"upvote-service/internal/db"
	"upvote-service/internal/domain"
	"upvote-service/internal/repository"
)

var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx, "postgres:17",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.WithSQLDriver("pgx"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("start postgres container: %v", err)
	}
	defer container.Terminate(ctx) //nolint:errcheck

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("get connection string: %v", err)
	}

	testPool, err = db.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect to test db: %v", err)
	}
	defer testPool.Close()

	if err := db.Migrate(testPool); err != nil {
		log.Fatalf("migrate test db: %v", err)
	}

	os.Exit(m.Run())
}

func truncate(t *testing.T) {
	t.Helper()
	_, err := testPool.Exec(context.Background(), "TRUNCATE TABLE upvotes")
	require.NoError(t, err)
}

func TestCreate_Success(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	input := domain.Upvote{ID: "018e1b0a-0000-7000-8000-000000000001", ArticleID: "a-1", UserID: "u-1"}
	got, err := repo.Create(context.Background(), input)

	require.NoError(t, err)
	assert.Equal(t, input.ID, got.ID)
	assert.Equal(t, input.ArticleID, got.ArticleID)
	assert.Equal(t, input.UserID, got.UserID)
	assert.False(t, got.CreatedAt.IsZero())
}

func TestCreate_DuplicateUpvote(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	input := domain.Upvote{ID: "018e1b0a-0000-7000-8000-000000000002", ArticleID: "a-1", UserID: "u-1"}
	_, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	// same user + article, different UUID
	input2 := domain.Upvote{ID: "018e1b0a-0000-7000-8000-000000000003", ArticleID: "a-1", UserID: "u-1"}
	_, err = repo.Create(context.Background(), input2)
	assert.ErrorIs(t, err, domain.ErrDuplicateUpvote)
}

func TestDelete_Success(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	input := domain.Upvote{ID: "018e1b0a-0000-7000-8000-000000000004", ArticleID: "a-1", UserID: "u-1"}
	_, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	err = repo.Delete(context.Background(), input.ID)
	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	err := repo.Delete(context.Background(), "018e1b0a-0000-7000-8000-000000000099")
	assert.Error(t, err)
}

func TestGetById_Success(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	input := domain.Upvote{ID: "018e1b0a-0000-7000-8000-000000000005", ArticleID: "a-1", UserID: "u-1"}
	_, err := repo.Create(context.Background(), input)
	require.NoError(t, err)

	got, err := repo.GetById(context.Background(), input.ID)
	require.NoError(t, err)
	assert.Equal(t, input.ID, got.ID)
	assert.Equal(t, input.ArticleID, got.ArticleID)
	assert.Equal(t, input.UserID, got.UserID)
}

func TestGetById_NotFound(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	_, err := repo.GetById(context.Background(), "018e1b0a-0000-7000-8000-000000000099")
	assert.Error(t, err)
}

func TestGetByArticle_ReturnsOnlyMatchingUpvotes(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	fixtures := []domain.Upvote{
		{ID: "018e1b0a-0000-7000-8000-000000000010", ArticleID: "a-1", UserID: "u-1"},
		{ID: "018e1b0a-0000-7000-8000-000000000011", ArticleID: "a-1", UserID: "u-2"},
		{ID: "018e1b0a-0000-7000-8000-000000000012", ArticleID: "a-2", UserID: "u-1"},
	}
	for _, f := range fixtures {
		_, err := repo.Create(context.Background(), f)
		require.NoError(t, err)
	}

	got, err := repo.GetByArticle(context.Background(), "a-1")
	require.NoError(t, err)
	assert.Len(t, got, 2)
	for _, u := range got {
		assert.Equal(t, "a-1", u.ArticleID)
	}
}

func TestGetByUser_ReturnsOnlyMatchingUpvotes(t *testing.T) {
	truncate(t)
	repo := repository.New(testPool)

	fixtures := []domain.Upvote{
		{ID: "018e1b0a-0000-7000-8000-000000000020", ArticleID: "a-1", UserID: "u-1"},
		{ID: "018e1b0a-0000-7000-8000-000000000021", ArticleID: "a-2", UserID: "u-1"},
		{ID: "018e1b0a-0000-7000-8000-000000000022", ArticleID: "a-1", UserID: "u-2"},
	}
	for _, f := range fixtures {
		_, err := repo.Create(context.Background(), f)
		require.NoError(t, err)
	}

	got, err := repo.GetByUser(context.Background(), "u-1")
	require.NoError(t, err)
	assert.Len(t, got, 2)
	for _, u := range got {
		assert.Equal(t, "u-1", u.UserID)
	}
}
