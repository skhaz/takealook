package router

import (
	"database/sql"
	_ "embed"
	"html/template"
	"io"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"skhaz.dev/urlshortnen/pkg/session"
)

var (
	validate = validator.New(validator.WithRequiredStructEnabled())

	//go:embed public/dashboard.html
	dashboard string

	//go:embed public/index.html
	index string

	//go:embed public/join.html
	join string

	//go:embed public/pay.html
	pay string

	//go:embed public/redirect.html
	redirect string

	//go:embed public/submit.html
	submit string
)

type TemplateRegistry struct {
	templates map[string]*template.Template
}

func (t *TemplateRegistry) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	tmpl, ok := t.templates[name]
	if !ok {
		return echo.NewHTTPError(500, "template not found")
	}

	return tmpl.Execute(w, data)
}

type Router struct {
	db        *sql.DB
	templates map[string]*template.Template
	echo      *echo.Echo
}

func NewRouter(db *sql.DB) *Router {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(middleware.GzipWithConfig(middleware.GzipConfig{MinLength: 3072}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(30)))
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())

	templates := make(map[string]*template.Template)
	templates["dashboard"] = template.Must(template.New("dashboard.html").Parse(dashboard))
	templates["index"] = template.Must(template.New("index.html").Parse(index))
	templates["join"] = template.Must(template.New("index.html").Parse(join))
	templates["pay"] = template.Must(template.New("pay.html").Parse(pay))
	templates["redirect"] = template.Must(template.New("redirect.html").Parse(redirect))
	templates["submit"] = template.Must(template.New("submit.html").Parse(submit))

	e.Renderer = &TemplateRegistry{
		templates: templates,
	}

	r := &Router{
		db:        db,
		templates: templates,
		echo:      e,
	}

	e.HEAD("/", r.Healthcheck)
	e.GET("/", r.Index)
	e.GET("/favicon.ico", r.Icon)
	e.GET("/robots.txt", r.Robots)
	e.Any("/submit", r.Submit, session.VerifyCookie)
	e.GET("/:short", r.Redirect)

	e.Any("/join", r.Join, session.SkipPassword)
	e.GET("/dashboard", r.Dashboard, session.VerifyCookie)
	e.GET("/pay", r.Pay, session.VerifyCookie)

	e.POST("/webhook", r.Webhook)

	return r
}

func (r *Router) Start(addr string) {
	r.echo.Logger.Fatal(r.echo.Start(addr))
}
