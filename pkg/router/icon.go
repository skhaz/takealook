package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Router) Icon(c echo.Context) error {
	return c.Blob(http.StatusOK, "image/x-icon", []byte{})
}
