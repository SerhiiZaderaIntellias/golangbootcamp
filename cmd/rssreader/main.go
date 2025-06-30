package main

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/db"
	rsshttp "github.com/SerhiiZaderaIntellias/golangbootcamp/internal/http"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	e := echo.New()

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		message := "Internal Server Error"

		// Extract status code and message from echo.HTTPError if available
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			if m, ok := he.Message.(string); ok {
				message = m
			} else {
				message = http.StatusText(code)
			}
		}

		// Log the error to console
		c.Logger().Error(err)

		// Return JSON response if nothing is sent yet
		if !c.Response().Committed {
			c.JSON(code, map[string]string{"error": message})
		}
	}

	handler := rsshttp.NewFeedHandler(database)

	e.POST("/feed", handler.CreateFeed)
	e.GET("/feed", handler.GetAllFeeds)
	e.GET("/feed/:id", handler.GetFeedByID)
	e.DELETE("/feed/:id", handler.DeleteFeed)

	e.Logger.Fatal(e.Start(":8080"))
}
