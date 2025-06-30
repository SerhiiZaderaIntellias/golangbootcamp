package rss

import (
	"encoding/xml"
	"time"
)

type Xml struct {
	XMLName xml.Name  `xml:"rss"`
	Channel []Channel `xml:"channel"`
}

type Channel struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Language    string `xml:"language"`
	Items       []Item `xml:"item"`
}

type Item struct {
	ID          int       `json:"id"`                // FOR DB
	Title       string    `json:"title" xml:"title"` // FOR XML
	Link        string    `json:"link" xml:"link"`
	Description string    `json:"description" xml:"description"`
	CreatedAt   time.Time `json:"created_at"` // FOR DB
}
