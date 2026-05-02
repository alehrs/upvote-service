CREATE TABLE IF NOT EXISTS upvotes (
    id         UUID        PRIMARY KEY,
    user_id    TEXT        NOT NULL,
    article_id TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_upvotes_user_article UNIQUE (user_id, article_id)
);

CREATE INDEX IF NOT EXISTS idx_upvotes_user_id    ON upvotes (user_id);
CREATE INDEX IF NOT EXISTS idx_upvotes_article_id ON upvotes (article_id);
