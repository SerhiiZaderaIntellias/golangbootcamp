# Makefile

DB_URL=postgres://rssuser:secret@localhost:5432/rssdb?sslmode=disable

run:
	go run ./cmd/rssreader

db-up:
	docker compose up -d

db-down:
	docker compose down

migrate-up:
	goose -dir ./migrations postgres "$(DB_URL)" up

migrate-down:
	goose -dir ./migrations postgres "$(DB_URL)" down

migrate-status:
	goose -dir ./migrations postgres "$(DB_URL)" status

migrate-create:
	goose -dir ./migrations create init_schema sql
