package rss

import (
	"database/sql"
	"fmt"
)

func StoreItems(db *sql.DB, items []Item) error {
	for _, item := range items {
		_, err := db.Exec(`
			INSERT INTO rss_items (title, link, description)
			VALUES ($1, $2, $3)
		`, item.Title, item.Link, item.Description)

		if err != nil {
			return err
		}
	}

	return nil
}

func GetFilteredFeeds(db *sql.DB, titleFilter, descFilter string, limit, offset int) ([]Item, error) {
	query := `
		SELECT id, title, link, description, created_at
		FROM rss_items
		WHERE 1=1
	`
	args := []interface{}{}
	argIdx := 1

	if titleFilter != "" {
		query += fmt.Sprintf(" AND title ILIKE $%d", argIdx)
		args = append(args, "%"+titleFilter+"%")
		argIdx++
	}

	if descFilter != "" {
		query += fmt.Sprintf(" AND description ILIKE $%d", argIdx)
		args = append(args, "%"+descFilter+"%")
		argIdx++
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Title, &item.Link, &item.Description, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func GetItemByID(db *sql.DB, id int) (*Item, error) {
	row := db.QueryRow(`
		SELECT id, title, link, description, created_at
		FROM rss_items
		WHERE id = $1
	`, id)

	var item Item
	err := row.Scan(&item.ID, &item.Title, &item.Link, &item.Description, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // not found
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func DeleteItemByID(db *sql.DB, id int) error {
	result, err := db.Exec(`DELETE FROM rss_items WHERE id = $1`, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
