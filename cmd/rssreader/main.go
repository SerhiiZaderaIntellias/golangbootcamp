package main

import (
	"fmt"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
)

func main() {
	url := "https://dou.ua/feed/"

	data, err := rss.FetchAndParse(url)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", data)
}
