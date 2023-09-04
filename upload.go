package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"cloud.google.com/go/storage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Upload(c echo.Context) error {
	return c.Render(http.StatusOK, "upload", nil)
}

func UploadFile(c echo.Context) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return c.String(500, err.Error())
	}
	defer client.Close()

	fhdr, err := c.FormFile("file")
	if err != nil {
		return c.String(500, err.Error())
	}
	f, err := fhdr.Open()
	if err != nil {
		return c.String(500, err.Error())
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	id := uuid.New().String()
	object := id + fhdr.Filename
	fmt.Printf("Object name: %s", object)
	o := client.Bucket("floo-network").Object(object)
	fmt.Printf("Bucket name: floo-network")
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return c.String(500, err.Error())
	}
	fmt.Println("Copy done")
	if err := wc.Close(); err != nil {
		fmt.Printf("Writer.Close: %s", err.Error())
		return c.String(500, err.Error())
	}
	fmt.Println("Writer closed")

	attrs, err := o.Attrs(ctx)
	fmt.Println("Attrs obtained")
	if err != nil {
		fmt.Printf("Object(%q).Attrs: %s", object, err.Error())
		return c.String(500, err.Error())
	}
	fmt.Printf("Media link: %s", attrs.MediaLink)

	return c.String(200, attrs.MediaLink)
}
