package router

import (
	_ "embed"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Router) Index(c echo.Context) error {
	return c.Render(http.StatusOK, "index", nil)
}
