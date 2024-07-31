package router

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Router) Healthcheck(c echo.Context) error {
	return c.NoContent(http.StatusOK)
}
