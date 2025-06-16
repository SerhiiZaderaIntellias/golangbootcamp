package rss

import (
	"database/sql"
)

func StoreItems(db *sql.DB, items []Item) error {
	for _, item := range items {
		_, err := db.Exec(`
			INSERT INTO rss_items (title, link, description) VALUES ($1, $2, $3)
		`, item.Title, item.Link, item.Description)
		if err != nil {
			return err
		}
	}
	return nil
}
