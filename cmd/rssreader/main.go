package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/labstack/echo/v4"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/db"
	rsshttp "github.com/SerhiiZaderaIntellias/golangbootcamp/internal/http"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/worker"
)

func main() {
	database, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer database.Close()

	// Start worker pool
	pool := worker.NewPool(database, 5)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
		<-quit
		log.Println("Received shutdown signal...")
		pool.Shutdown()
		os.Exit(0)
	}()

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

	handler := rsshttp.NewFeedHandler(database, pool)

	e.POST("/feed", handler.CreateFeed)
	e.GET("/feed", handler.GetAllFeeds)
	e.GET("/feed/:id", handler.GetFeedByID)
	e.DELETE("/feed/:id", handler.DeleteFeed)

	log.Println("Server running on :8080")
	e.Logger.Fatal(e.Start(":8080"))
}
