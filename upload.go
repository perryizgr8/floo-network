package main

import (
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Upload(c echo.Context) error {
	return c.Render(http.StatusOK, "upload", nil)
}

func UploadFile(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return err
	}
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()
	buf := make([]byte, 1024)
	for {
		// read a chunk
		n, err := src.Read(buf)
		if err != nil && err != io.EOF {
			panic(err)
		}
		if n == 0 {
			break
		}
	}
	return c.String(200, string(buf[:]))
}
