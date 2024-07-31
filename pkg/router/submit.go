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

func (r *Router) Submit(c echo.Context) error {
	email, ok := c.Get("email").(string)
	if !ok {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "email is not present or is not a string"})
	}

	var (
		message string
		err     error
	)

	switch c.Request().Method {
	case http.MethodGet:
		return c.Render(http.StatusOK, "submit", struct{ CDN string }{CDN: cdn})
	case http.MethodPost:
		type Payload struct {
			Url string `json:"url" validate:"required,url"`
		}

		var payload Payload
		if err = c.Bind(&payload); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "failed to bind data"})
		}

		if err = validate.Struct(payload); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "validation failed"})
		}

		var id int64
		var error sql.NullString
		var ready bool
		var title string
		var description string
		if err = r.db.QueryRow("SELECT id, error, ready, title, description FROM data WHERE url = ? LIMIT 1", payload.Url).Scan(&id, &error, &ready, &title, &description); err != nil {
			if err == sql.ErrNoRows {
				result, err := r.db.Exec(`
					INSERT INTO data (url, user)
					SELECT d.url, d.user
					FROM (SELECT ? AS url, ? AS user) AS d
					JOIN users u ON u.email = d.user AND u.active = 1
				`, payload.Url, email)
				if err != nil {
					message = "database insert failed"
					log.Error(message, zap.Error(err))
					return c.JSON(http.StatusInternalServerError, echo.Map{"error": message})
				}

				if affectedRows, _ := result.RowsAffected(); affectedRows == 0 {
					message = "no data inserted, possibly due to unpaid user"
					log.Warn(message)
					return c.JSON(http.StatusPaymentRequired, echo.Map{})
				}
			}
		}

		if ready {
			if error.Valid {
				return c.JSON(http.StatusBadRequest, echo.Map{"error": error.String})
			}

			if id < 0 {
				message = "invalid id: id must be non-negative"
				log.Warn(message)
				return c.JSON(http.StatusBadRequest, echo.Map{"error": message})
			}

			var (
				short   = base36.Encode(uint64(id))
				preview = fmt.Sprintf("%s/%s.webp", cdn, short)
				url     = fmt.Sprintf("%s/%s", domain, short)
			)

			return c.JSON(http.StatusCreated, echo.Map{"preview": preview, "title": title, "description": description, "url": url})
		}

		return c.JSON(http.StatusTooManyRequests, echo.Map{"error": "too many requests"})
	default:
		return c.JSON(http.StatusMethodNotAllowed, echo.Map{"error": "method not allowed"})
	}
}
