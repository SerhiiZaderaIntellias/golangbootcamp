package rss

import (
	"fmt"
	"io"
	"net/http"
)

func Fetch(url string) ([]byte, error) {
	res, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	return io.ReadAll(res.Body)
}
