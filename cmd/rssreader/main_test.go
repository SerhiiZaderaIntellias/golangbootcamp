package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/db"
	rsshttp "github.com/SerhiiZaderaIntellias/golangbootcamp/internal/http"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/worker"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
)

func setupTestServer(t *testing.T) (*echo.Echo, *sql.DB, func()) {
	// Set test environment
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "rssuser")
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("DB_NAME", "rssdb")

	// Connect to test database
	database, err := db.Connect()
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
	}

	// Clear test table (don't recreate it since it already exists)
	_, err = database.Exec(`DELETE FROM rss_items_test`)
	if err != nil {
		t.Skipf("Skipping test: cannot clear test table: %v", err)
	}

	// Setup Echo server
	e := echo.New()

	// Start worker pool
	pool := worker.NewPool(database, 2)

	// Create test handler that uses test table
	testHandler := &testFeedHandler{
		db:   database,
		pool: pool,
	}

	// Setup routes
	e.POST("/feed", testHandler.CreateFeed)
	e.GET("/feed", testHandler.GetAllFeeds)
	e.GET("/feed/:id", testHandler.GetFeedByID)
	e.DELETE("/feed/:id", testHandler.DeleteFeed)

	cleanup := func() {
		pool.Shutdown()
		database.Close()
	}

	return e, database, cleanup
}

// Test versions of functions that work with rss_items_test table
func getFilteredFeedsTest(db *sql.DB, titleFilter, descFilter string, limit, offset int) ([]rss.Item, error) {
	query := `
		SELECT id, title, link, description, created_at
		FROM rss_items_test
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

	var items []rss.Item
	for rows.Next() {
		var item rss.Item
		err := rows.Scan(&item.ID, &item.Title, &item.Link, &item.Description, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func getItemByIDTest(db *sql.DB, id int) (*rss.Item, error) {
	row := db.QueryRow(`
		SELECT id, title, link, description, created_at
		FROM rss_items_test
		WHERE id = $1
	`, id)

	var item rss.Item
	err := row.Scan(&item.ID, &item.Title, &item.Link, &item.Description, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil // not found
	}

	if err != nil {
		return nil, err
	}

	return &item, nil
}

func deleteItemByIDTest(db *sql.DB, id int) error {
	result, err := db.Exec(`DELETE FROM rss_items_test WHERE id = $1`, id)
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

// Test handler that uses test table
type testFeedHandler struct {
	db   *sql.DB
	pool *worker.Pool
}

func (h *testFeedHandler) CreateFeed(c echo.Context) error {
	var req rsshttp.FeedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	h.pool.Submit(req.URL)

	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h *testFeedHandler) GetAllFeeds(c echo.Context) error {
	title := c.QueryParam("title")
	description := c.QueryParam("description")

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, err := getFilteredFeedsTest(h.db, title, description, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, items)
}

func (h *testFeedHandler) GetFeedByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}

	item, err := getItemByIDTest(h.db, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	return c.JSON(http.StatusOK, item)
}

func (h *testFeedHandler) DeleteFeed(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}

	err = deleteItemByIDTest(h.db, id)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent) // 204
}

func TestCreateFeed_Integration(t *testing.T) {
	e, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Create test RSS server
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

	// Test POST /feed
	reqBody := map[string]string{"url": server.URL}
	reqJSON, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/feed", bytes.NewBuffer(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("Expected status %d, got %d", http.StatusAccepted, rec.Code)
	}

	// Wait for worker to process
	time.Sleep(1 * time.Second)

	// Verify item was stored
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM rss_items_test").Scan(&count)
	if err != nil {
		t.Errorf("Failed to count items: %v", err)
	}
	if count == 0 {
		t.Error("Expected items to be stored")
	}
}

func TestGetAllFeeds_Integration(t *testing.T) {
	e, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Insert test data
	_, err := db.Exec(`
		INSERT INTO rss_items_test (title, link, description)
		VALUES ($1, $2, $3), ($1, $4, $3)
	`, "Test Item", "http://example.com/item1", "Test Description", "http://example.com/item2")
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test GET /feed
	req := httptest.NewRequest(http.MethodGet, "/feed", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test with filters
	req = httptest.NewRequest(http.MethodGet, "/feed?title=Test&limit=1", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestGetFeedByID_Integration(t *testing.T) {
	e, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Insert test data
	var id int
	err := db.QueryRow(`
		INSERT INTO rss_items_test (title, link, description)
		VALUES ($1, $2, $3) RETURNING id
	`, "Test Item", "http://example.com/item", "Test Description").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test GET /feed/:id
	req := httptest.NewRequest(http.MethodGet, "/feed/"+fmt.Sprintf("%d", id), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test invalid ID
	req = httptest.NewRequest(http.MethodGet, "/feed/abc", nil)
	rec = httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteFeed_Integration(t *testing.T) {
	e, db, cleanup := setupTestServer(t)
	defer cleanup()

	// Insert test data
	var id int
	err := db.QueryRow(`
		INSERT INTO rss_items_test (title, link, description)
		VALUES ($1, $2, $3) RETURNING id
	`, "Test Item", "http://example.com/item", "Test Description").Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	// Test DELETE /feed/:id
	req := httptest.NewRequest(http.MethodDelete, "/feed/"+fmt.Sprintf("%d", id), nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
	}

	// Verify item was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM rss_items_test WHERE id = $1", id).Scan(&count)
	if err != nil {
		t.Errorf("Failed to count items: %v", err)
	}
	if count != 0 {
		t.Error("Expected item to be deleted")
	}
}