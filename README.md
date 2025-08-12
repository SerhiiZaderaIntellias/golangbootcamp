# RSS Reader

A simple RSS reader application built with Go and Echo framework.

## Features

- Fetch and parse RSS feeds
- Store RSS items in PostgreSQL database
- REST API endpoints for managing feeds
- Concurrent processing with worker pools
- Graceful shutdown support

## API Endpoints

- `POST /feed` - Add new RSS feed URL
- `GET /feed` - Get all feeds with optional filters
- `GET /feed/:id` - Get specific feed by ID
- `DELETE /feed/:id` - Delete specific feed by ID

## Setup

1. Start the app:
```bash
make docker-up
```

2. Run migrations:
```bash
make migrate-up
```

3. Run the application:
```bash
make run
```

## Testing

The project has comprehensive test coverage including:

### Unit Tests
- `pkg/rss/` - RSS parsing and database operations
- `internal/worker/` - Worker pool functionality
- `internal/http/` - HTTP handlers

### Integration Tests
- `cmd/rssreader/` - Full application integration tests

### Running Tests

```bash
# Run all tests
make test

# Run with coverage report
make test-coverage

# Run with race detection
make test-race

# Run benchmarks
make test-benchmark

# Run short tests only
make test-short
```

### Test Coverage

The project focuses on testing **business logic and flow** rather than implementation details:

```
pkg/rss/
├── fetch_test.go      # HTTP fetching
├── parse_test.go      # XML parsing
└── integration_test.go # End-to-end flow

internal/worker/
└── pool_test.go       # Worker pool flow

internal/http/
└── handler_test.go    # HTTP handlers

cmd/rssreader/
└── main_test.go       # Full app integration
```

### Test Philosophy

We test **what** the code does, not **how** it does it:

- **Flow Testing** - Test complete user journeys (URL → RSS → Parse → Store)
- **Error Handling** - Test how system handles failures gracefully
- **Integration** - Test components working together
- **Mocking** - Mock external dependencies (DB, HTTP) to focus on business logic

## Architecture

The application uses a worker pool pattern for concurrent RSS processing:
```
HTTP POST /feed
    ↓
handler.CreateFeed()
    ↓
h.pool.Submit(req.URL)
    ↓
p.jobs <- url (channel)
    ↓
[QUEUE: url1, url2, url3...]
    ↓
case url := <-p.jobs (worker receives)
    ↓
rss.FetchAndParse(url)
    ↓
rss.StoreItems(db, items)
    ↓
✅ DONE!
```

## Graceful Shutdown

The application implements graceful shutdown:
1. Receives SIGINT/SIGTERM signal
2. Cancels worker context
3. Waits for all workers to complete
4. Closes database connections
5. Exits cleanly