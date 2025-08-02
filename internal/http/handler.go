package http

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/SerhiiZaderaIntellias/golangbootcamp/pkg/rss"
	"github.com/SerhiiZaderaIntellias/golangbootcamp/internal/worker"
	"github.com/labstack/echo/v4"
)

type FeedHandler struct {
	db   *sql.DB
	pool *worker.Pool
}

func NewFeedHandler(db *sql.DB, pool *worker.Pool) *FeedHandler {
	return &FeedHandler{db: db, pool: pool}
}

type FeedRequest struct {
	URL string `json:"url"`
}

func (h *FeedHandler) CreateFeed(c echo.Context) error {
	var req FeedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	h.pool.Submit(req.URL)

	return c.JSON(http.StatusAccepted, map[string]string{"status": "queued"})
}

func (h *FeedHandler) GetAllFeeds(c echo.Context) error {
	title := c.QueryParam("title")
	description := c.QueryParam("description")

	limit, err := strconv.Atoi(c.QueryParam("limit"))
	if err != nil || limit <= 0 {
		limit = 10
	}

	offset, err := strconv.Atoi(c.QueryParam("offset"))
	if err != nil || offset < 0 {
		offset = 0
	}

	items, err := rss.GetFilteredFeeds(h.db, title, description, limit, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, items)
}

func (h *FeedHandler) GetFeedByID(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}

	item, err := rss.GetItemByID(h.db, id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	if item == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	return c.JSON(http.StatusOK, item)
}

func (h *FeedHandler) DeleteFeed(c echo.Context) error {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid ID"})
	}

	err = rss.DeleteItemByID(h.db, id)
	if err == sql.ErrNoRows {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent) // 204
}
