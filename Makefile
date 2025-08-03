# Makefile

DB_URL=postgres://rssuser:secret@localhost:5432/rssdb?sslmode=disable

run:
	go run ./cmd/rssreader

docker-up:
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

test:
	go test ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

test-race:
	go test -race ./...

test-benchmark:
	go test -bench=. ./...

test-short:
	go test -short ./...
