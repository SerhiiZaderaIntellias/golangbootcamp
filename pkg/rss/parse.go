package rss

import (
	"encoding/xml"
)

func Parse(data []byte) (*Xml, error) {
	var result Xml
	err := xml.Unmarshal(data, &result)
	return &result, err
}

func FetchAndParse(url string) (*Xml, error) {
	b, err := Fetch(url)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}
