package router

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/martinlindhe/base36"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
)

func (r *Router) Dashboard(c echo.Context) error {
	email, ok := c.Get("email").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "email is not present or is not a string"})
	}

	var offset int64

	type Data struct {
		ID    int
		Count int
		URL   string
		Title string
		Short string
		Error string
		Ready int
	}

	var result []Data

	rows, err := r.db.Query(`
		SELECT id, count, url, title, error, ready
		FROM data
		WHERE user = ?
		ORDER BY created_at
		LIMIT 10 OFFSET ?
	`, email, offset)
	if err != nil {
		var message = "failed to query data"
		log.Error(message, zap.Error(err))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": message})
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		var count int
		var url, title, error sql.NullString
		var ready int

		if err := rows.Scan(&id, &count, &url, &title, &error, &ready); err != nil {
			log.Error("failed to scan row", zap.Error(err))
			continue
		}

		if id < 0 {
			var message = "invalid id: id must be non-negative"
			log.Warn(message)
			return c.JSON(http.StatusBadRequest, echo.Map{"error": message})
		}

		result = append(result, Data{
			ID:    id,
			Count: count,
			URL:   url.String,
			Title: title.String,
			Short: fmt.Sprintf("%s/%s", domain, base36.Encode(uint64(id))),
			Error: error.String,
			Ready: ready,
		})
	}

	if err := rows.Err(); err != nil {
		var message = "error occurred during row iteration"
		log.Error(message, zap.Error(err))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": message})
	}

	return c.Render(http.StatusOK, "dashboard", struct{ Data []Data }{Data: result})
}
