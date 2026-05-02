package domain

import (
	"errors"
	"time"
)

var ErrDuplicateUpvote = errors.New("user already upvoted this article")

type Upvote struct {
	ID        string
	UserID    string
	ArticleID string
	CreatedAt time.Time
}
