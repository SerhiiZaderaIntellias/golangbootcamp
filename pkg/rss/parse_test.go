package rss

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name        string
		xmlData     string
		expectError bool
		expected    *Xml
	}{
		{
			name: "valid RSS",
			xmlData: `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Channel</title>
		<link>http://example.com</link>
		<description>Test Description</description>
		<language>en</language>
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
</rss>`,
			expectError: false,
			expected: &Xml{
				Channel: []Channel{
					{
						Title:       "Test Channel",
						Link:        "http://example.com",
						Description: "Test Description",
						Language:    "en",
						Items: []Item{
							{
								Title:       "Test Item 1",
								Link:        "http://example.com/item1",
								Description: "Test Item 1 Description",
							},
							{
								Title:       "Test Item 2",
								Link:        "http://example.com/item2",
								Description: "Test Item 2 Description",
							},
						},
					},
				},
			},
		},
		{
			name:        "invalid XML",
			xmlData:     "<invalid>xml</invalid>",
			expectError: true,
		},
		{
			name:        "empty XML",
			xmlData:     "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parse([]byte(tt.xmlData))

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if result == nil {
					t.Error("Expected result but got nil")
					return
				}
				if len(result.Channel) != len(tt.expected.Channel) {
					t.Errorf("Expected %d channels, got %d", len(tt.expected.Channel), len(result.Channel))
				}
				if len(result.Channel) > 0 {
					channel := result.Channel[0]
					expectedChannel := tt.expected.Channel[0]
					if channel.Title != expectedChannel.Title {
						t.Errorf("Expected title %s, got %s", expectedChannel.Title, channel.Title)
					}
					if channel.Link != expectedChannel.Link {
						t.Errorf("Expected link %s, got %s", expectedChannel.Link, channel.Link)
					}
					if len(channel.Items) != len(expectedChannel.Items) {
						t.Errorf("Expected %d items, got %d", len(expectedChannel.Items), len(channel.Items))
					}
				}
			}
		})
	}
}

func TestFetchAndParse(t *testing.T) {
	// Create test server with valid RSS
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
	<channel>
		<title>Test Channel</title>
		<item>
			<title>Test Item</title>
			<link>http://example.com/item</link>
		</item>
	</channel>
</rss>`))
	}))
	defer server.Close()

	result, err := FetchAndParse(server.URL)
	if err != nil {
		t.Errorf("Expected no error but got: %v", err)
	}
	if result == nil {
		t.Error("Expected result but got nil")
	}
	if len(result.Channel) == 0 {
		t.Error("Expected channel but got none")
	}
}