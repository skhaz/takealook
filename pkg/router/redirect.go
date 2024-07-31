package router

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/martinlindhe/base36"
	"github.com/mileusna/useragent"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
	. "skhaz.dev/urlshortnen/pkg/functions"
)

var (
	cdn = os.Getenv("BUNNY_CDN")

	domain = os.Getenv("DOMAIN")

	extension = "webp"
)

func (r *Router) Redirect(c echo.Context) error {
	var (
		url         string
		title       string
		description string
		short       = c.Param("short")
		id          = base36.Decode(short)
		query       = "SELECT url, title, description FROM data WHERE id = ? LIMIT 1"
		err         = r.db.QueryRow(query, id).Scan(&url, &title, &description)
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return c.String(http.StatusNotFound, "not found")
		}

		var message = fmt.Sprintf("database query failed for '%s'", short)
		log.Error(message, zap.Error(err))
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": message})
	}

	if ua := useragent.Parse(c.Request().Header.Get("User-Agent")); !ua.Bot {
		Increment(id)
	}

	arguments := struct {
		Title       string
		Description string
		Image       string
		URL         string
		Canonical   string
	}{
		Title:       title,
		Description: description,
		Image:       fmt.Sprintf("%s/%s.%s", cdn, short, extension),
		URL:         url,
		Canonical:   fmt.Sprintf("%s/%s", domain, short),
	}

	return c.Render(http.StatusOK, "redirect", arguments)
}
