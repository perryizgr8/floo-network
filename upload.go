package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UploadData struct {
	Url          string
	TempFilename string
}

type RenameData struct {
	Filename     string `json:"filename"`
	TempFilename string `json:"tempfilename"`
}

func Upload(c echo.Context) error {
	secret := c.FormValue("secret")
	if secret != os.Getenv("SECRET") {
		return c.String(http.StatusUnauthorized, "Perhaps your memory fails you.")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer client.Close()

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "PUT",
		Headers: []string{"Content-Type:application/octet-stream"},
		Expires: time.Now().Add(15 * time.Minute),
	}

	id := uuid.New().String()
	object := fmt.Sprintf("%s/%s", id, "temp")
	u, err := client.Bucket("floo-transit").SignedURL(object, opts)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.Render(http.StatusOK, "upload", UploadData{Url: u, TempFilename: object})
}

func RenameAndGetDownloadUrl(c echo.Context) error {
	renameData := new(RenameData)
	if err := c.Bind(renameData); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	filename := renameData.Filename
	tempFilename := renameData.TempFilename

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer client.Close()

	// Rename the file from temp to the actual filename
	id := uuid.New().String()
	object := fmt.Sprintf("%s/%s", id, filename)
	_, err = client.Bucket("floo-transit").Object(object).CopierFrom(client.Bucket("floo-transit").Object(tempFilename)).Run(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if err := client.Bucket("floo-transit").Object(tempFilename).Delete(ctx); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(48 * time.Hour),
	}

	u, err := client.Bucket("floo-transit").SignedURL(object, opts)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	fireclnt, err := firestore.NewClient(ctx, "floo-network")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer fireclnt.Close()

	_, _, err = fireclnt.Collection("files").Add(ctx, map[string]interface{}{
		"uuid":       id,
		"signed_url": u,
	})
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("https://floo.perryizgr8.com/accio/%s", id))
}
