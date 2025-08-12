package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRSSFlow_EndToEnd(t *testing.T) {
	// Create test RSS server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Channel</title>
		<link>http://example.com</link>
		<description>Test Description</description>
		<item>
			<title>Test Item 1</title>
			<link>http://example.com/item1</link>
			<description>Test Item 1 Description</description>
		</item>
		<item>
			<title>Test Item 2</title>
			<link>http://example.com/item2</link>
			<description>Test Item 2 Description</description>
		</item>
	</channel>
</rss>`))
	}))
	defer server.Close()

	// Test complete flow: Fetch -> Parse
	rssData, err := FetchAndParse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	if rssData == nil {
		t.Fatal("Expected RSS data but got nil")
	}

	if len(rssData.Channel) == 0 {
		t.Fatal("Expected channel but got none")
	}

	channel := rssData.Channel[0]
	if channel.Title != "Test Channel" {
		t.Errorf("Expected title 'Test Channel', got '%s'", channel.Title)
	}

	if len(channel.Items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(channel.Items))
	}

	// Verify first item
	item1 := channel.Items[0]
	if item1.Title != "Test Item 1" {
		t.Errorf("Expected title 'Test Item 1', got '%s'", item1.Title)
	}
	if item1.Link != "http://example.com/item1" {
		t.Errorf("Expected link 'http://example.com/item1', got '%s'", item1.Link)
	}
}

func TestRSSFlow_ErrorHandling(t *testing.T) {
	// Test invalid RSS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<invalid>xml</invalid>`))
	}))
	defer server.Close()

	_, err := FetchAndParse(server.URL)
	if err == nil {
		t.Error("Expected error for invalid XML")
	}
}

func TestRSSFlow_HTTPErrors(t *testing.T) {
	// Test server error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	_, err := FetchAndParse(server.URL)
	if err == nil {
		t.Error("Expected error for 404 status")
	}
}

func TestRSSFlow_EmptyRSS(t *testing.T) {
	// Test empty RSS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Empty Channel</title>
	</channel>
</rss>`))
	}))
	defer server.Close()

	rssData, err := FetchAndParse(server.URL)
	if err != nil {
		t.Fatalf("Expected no error but got: %v", err)
	}

	if len(rssData.Channel) == 0 {
		t.Fatal("Expected channel but got none")
	}

	channel := rssData.Channel[0]
	if len(channel.Items) != 0 {
		t.Errorf("Expected 0 items, got %d", len(channel.Items))
	}
}