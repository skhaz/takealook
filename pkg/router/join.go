package router

import (
	"database/sql"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	log "skhaz.dev/urlshortnen/logging"
	"skhaz.dev/urlshortnen/pkg/mailer"
	"skhaz.dev/urlshortnen/pkg/session"
)

type Form struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=32"`
}

func (r *Router) Join(c echo.Context) error {
	arguments := struct {
		Error   bool
		Message string
	}{}

	switch c.Request().Method {
	case http.MethodGet:
	case http.MethodPost:
		var (
			form = Form{
				Email:    c.FormValue("email"),
				Password: c.FormValue("password"),
			}
			password string
			equals   bool
			salt     []byte
			hash     string
			err      error
		)

		if err = validate.Struct(form); err != nil {
			arguments.Error = true
			arguments.Message = fmt.Sprintf("validation failed %v", err)
			log.Error(arguments.Message, zap.Error(err))
			break
		}

		err = r.db.QueryRow("SELECT password FROM users WHERE email = ? LIMIT 1", form.Email).Scan(&password)

		switch {
		case err == sql.ErrNoRows:
			if salt, err = session.GenerateSalt(); err != nil {
				arguments.Error = true
				arguments.Message = "error generating salt for password"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			if hash, err = session.HashPassword(form.Password, salt); err != nil {
				arguments.Error = true
				arguments.Message = "error hashing password"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			if _, err = r.db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", form.Email, hash); err != nil {
				arguments.Error = true
				arguments.Message = "error inserting user into database"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			//nolint:golint,errcheck
			go mailer.NewMail().Send("rodrigo@delduca.org", "New user", form.Email)

			if err = session.SetCookie(form.Email, c); err != nil {
				arguments.Error = true
				arguments.Message = "error setting cookie"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			return c.Redirect(http.StatusFound, "/dashboard")
		case err != nil:
			arguments.Error = true
			arguments.Message = "error querying user"
			log.Error(arguments.Message, zap.Error(err))
		default:
			if equals, err = session.ComparePasswords(password, form.Password); !equals || err != nil {
				arguments.Error = true
				arguments.Message = "password mismatch"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			if err = session.SetCookie(form.Email, c); err != nil {
				arguments.Error = true
				arguments.Message = "error setting cookie"
				log.Error(arguments.Message, zap.Error(err))
				break
			}

			return c.Redirect(http.StatusFound, "/dashboard")
		}

	default:
		return c.JSON(http.StatusMethodNotAllowed, echo.Map{"error": "method not allowed"})
	}

	return c.Render(http.StatusOK, "join", arguments)
}
