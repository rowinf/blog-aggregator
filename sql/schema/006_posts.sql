-- +goose Up
CREATE TABLE posts(
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title TEXT NOT NULL,
    url TEXT UNIQUE NOT NULL,
    description TEXT NOT NULL,
    published_at TIMESTAMP NOT NULL,
    feed_id TEXT REFERENCES feeds (id) ON DELETE CASCADE NOT NULL
);

-- +goose Down
DROP TABLE posts;