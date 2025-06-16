package main

import (
	"fmt"
	"log"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/db"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
)

func main() {
	url := "https://dou.ua/feed/"

	rssData, err := rss.FetchAndParse(url)

	if err != nil {
		log.Fatal(err)
	}

	db, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = rss.StoreItems(db, rssData.Channel[0].Items)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("RSS data saved successfully!")
}
