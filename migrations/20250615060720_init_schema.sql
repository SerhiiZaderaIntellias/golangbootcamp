-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS rss_items (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    link TEXT,
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS rss_items;
-- +goose StatementEnd
