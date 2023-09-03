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
		return fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	fhdr, err := c.FormFile("file")
	if err != nil {
		return err
	}
	f, err := fhdr.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	id := uuid.New().String()
	object := id + fhdr.Filename
	o := client.Bucket("floo-network").Object(object)
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %w", err)
	}

	attrs, err := o.Attrs(ctx)
	if err != nil {
		return fmt.Errorf("Object(%q).Attrs: %w", object, err)
	}

	return c.String(200, attrs.MediaLink)
}
