-- +goose Up
CREATE TABLE feed_follows(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    user_id TEXT REFERENCES users (id) ON DELETE CASCADE NOT NULL,
    feed_id TEXT REFERENCES feeds (id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE feed_follows;