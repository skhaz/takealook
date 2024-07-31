package router

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (r *Router) Activate(c echo.Context) error {
	query := `UPDATE users SET active = 1 WHERE uuid = ?`

	if _, err := r.db.Exec(query, c.Param("uuid")); err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.HTML(http.StatusOK, "Thank you!")
}
