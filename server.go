package main

import (
	"html/template"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func main() {
	e := echo.New()
	t := &Template{templates: template.Must(template.ParseFiles("upload.html"))}
	e.Renderer = t
	e.Logger.SetLevel(log.DEBUG)
	e.GET("/", Upload)
	e.POST("upload", UploadFile)
	e.GET("health/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	e.Logger.Fatal(e.Start(":1323"))
}
