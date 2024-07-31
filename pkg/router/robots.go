package router

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func (r *Router) Robots(c echo.Context) error {
	return c.String(http.StatusOK, strings.Join([]string{"user-agent: *", "disallow:"}, "\n"))
}
