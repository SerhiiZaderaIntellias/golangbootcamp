package worker

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
)

// Mock DB for testing
type mockDB struct {
	items []rss.Item
	mu    sync.Mutex
}

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simulate storing items
	if len(args) >= 3 {
		item := rss.Item{
			Title:       args[0].(string),
			Link:        args[1].(string),
			Description: args[2].(string),
		}
		m.items = append(m.items, item)
	}

	return &mockResult{}, nil
}

// Create test pool with mock DB
func createTestPool() *Pool {
	ctx, cancel := context.WithCancel(context.Background())
	return &Pool{
		db:     nil, // We'll mock the actual DB operations
		jobs:   make(chan string, 100),
		ctx:    ctx,
		cancel: cancel,
	}
}

type mockResult struct{}

func (m *mockResult) LastInsertId() (int64, error) { return 1, nil }
func (m *mockResult) RowsAffected() (int64, error) { return 1, nil }

func TestNewPool(t *testing.T) {
	pool := createTestPool()

	if pool == nil {
		t.Error("Expected pool to be created")
	}
	if cap(pool.jobs) != 100 {
		t.Errorf("Expected jobs channel capacity 100, got %d", cap(pool.jobs))
	}
}

func TestPool_Submit(t *testing.T) {
	pool := createTestPool()

	// Test successful submission
	pool.Submit("http://example.com/rss1")

	// Test queue full (we need to fill the buffer)
	for i := 0; i < 101; i++ {
		pool.Submit("http://example.com/rss")
	}

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)
}

func TestPool_Shutdown(t *testing.T) {
	pool := createTestPool()

	// Submit some jobs
	pool.Submit("http://example.com/rss1")
	pool.Submit("http://example.com/rss2")

	// Shutdown should complete without hanging
	done := make(chan bool)
	go func() {
		pool.Shutdown()
		done <- true
	}()

	select {
	case <-done:
		// Shutdown completed successfully
	case <-time.After(2 * time.Second):
		t.Error("Shutdown timed out")
	}
}

func TestPool_WorkerProcessing(t *testing.T) {
	pool := createTestPool()

	// Create a test server that returns valid RSS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Channel</title>
		<item>
			<title>Test Item</title>
			<link>http://example.com/item</link>
			<description>Test Description</description>
		</item>
	</channel>
</rss>`))
	}))
	defer server.Close()

	// Submit job
	pool.Submit(server.URL)

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Shutdown
	pool.Shutdown()
}

func TestPool_WorkerErrorHandling(t *testing.T) {
	pool := createTestPool()

	// Submit invalid URL
	pool.Submit("http://invalid-url-that-does-not-exist.com")

	// Wait for processing
	time.Sleep(500 * time.Millisecond)

	// Shutdown
	pool.Shutdown()
}

func TestPool_ContextCancellation(t *testing.T) {
	pool := createTestPool()

	// Submit jobs
	pool.Submit("http://example.com/rss1")
	pool.Submit("http://example.com/rss2")

	// Shutdown should complete quickly
	done := make(chan bool)
	go func() {
		pool.Shutdown()
		done <- true
	}()

	select {
	case <-done:
		// Shutdown completed successfully
	case <-time.After(1 * time.Second):
		t.Error("Shutdown timed out")
	}
}