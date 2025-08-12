-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rss_items_test (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    link TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_rss_items_test_link ON rss_items_test (link);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rss_items_test;
-- +goose StatementEnd