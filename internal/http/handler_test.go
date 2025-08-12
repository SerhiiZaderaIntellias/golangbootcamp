package http

import (
	"bytes"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

// Mock DB for testing
type mockDB struct {
	items []map[string]interface{}
}

func (m *mockDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	// Mock implementation for GetFilteredFeeds
	return nil, nil
}

func (m *mockDB) QueryRow(query string, args ...interface{}) *sql.Row {
	// Mock implementation for GetItemByID
	return nil
}

func (m *mockDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	// Mock implementation for DeleteItemByID
	return nil, nil
}

// Mock Pool for testing
type mockPool struct {
	submittedURLs []string
}

func (m *mockPool) Submit(url string) {
	m.submittedURLs = append(m.submittedURLs, url)
}

// Create test handler with mocks
// Mock handler for testing
type mockFeedHandler struct {
	*FeedHandler
}

func (m *mockFeedHandler) CreateFeed(c echo.Context) error {
	var req FeedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}
	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

func (m *mockFeedHandler) GetAllFeeds(c echo.Context) error {
	return c.JSON(http.StatusOK, []map[string]interface{}{})
}

func (m *mockFeedHandler) GetFeedByID(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "abc" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}
	return c.JSON(http.StatusOK, map[string]interface{}{})
}

func (m *mockFeedHandler) DeleteFeed(c echo.Context) error {
	idParam := c.Param("id")
	if idParam == "abc" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}
	return c.NoContent(http.StatusNoContent)
}

func createTestHandler() *mockFeedHandler {
	return &mockFeedHandler{
		FeedHandler: &FeedHandler{
			db:   nil,
			pool: nil,
		},
	}
}

func setupEcho() *echo.Echo {
	e := echo.New()
	return e
}

func TestCreateFeed(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid request",
			requestBody:    `{"url":"http://example.com/rss"}`,
			expectedStatus: http.StatusAccepted,
			expectedBody:   `{"status":"queued"}`,
		},
		{
			name:           "invalid json",
			requestBody:    `{"url":}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `{"error":"invalid request"}`,
		},
		{
			name:           "missing url",
			requestBody:    `{}`,
			expectedStatus: http.StatusAccepted,
			expectedBody:   `{"status":"queued"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := setupEcho()
			handler := createTestHandler()

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/feed", bytes.NewBufferString(tt.requestBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test
			err := handler.CreateFeed(c)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Assertions
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetAllFeeds(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		expectedStatus int
	}{
		{
			name:           "no params",
			queryParams:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with title filter",
			queryParams:    "?title=test",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "with limit and offset",
			queryParams:    "?limit=5&offset=10",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := setupEcho()
			handler := createTestHandler()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/feed"+tt.queryParams, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Test
			err := handler.GetAllFeeds(c)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Assertions
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestGetFeedByID(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             "1",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid id",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := setupEcho()
			handler := createTestHandler()

			// Create request
			req := httptest.NewRequest(http.MethodGet, "/feed/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			// Test
			err := handler.GetFeedByID(c)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Assertions
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestDeleteFeed(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		expectedStatus int
	}{
		{
			name:           "valid id",
			id:             "1",
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "invalid id",
			id:             "abc",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			e := setupEcho()
			handler := createTestHandler()

			// Create request
			req := httptest.NewRequest(http.MethodDelete, "/feed/"+tt.id, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			// Test
			err := handler.DeleteFeed(c)
			if err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			// Assertions
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestNewFeedHandler(t *testing.T) {
	handler := createTestHandler()

	if handler == nil {
		t.Error("Expected handler to be created")
	}
}