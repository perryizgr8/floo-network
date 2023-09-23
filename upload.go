package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func Upload(c echo.Context) error {
	secret := c.FormValue("secret")
	if secret != os.Getenv("SECRET") {
		return c.String(http.StatusUnauthorized, "Perhaps your memory fails you.")
	}

	// ctx := context.Background()
	// client, err := storage.NewClient(ctx)
	// if err != nil {
	// 	return c.String(http.StatusInternalServerError, err.Error())
	// }

	return c.Render(http.StatusOK, "upload", nil)
}

func UploadFile(c echo.Context) error {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer client.Close()

	fhdr, err := c.FormFile("file")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	f, err := fhdr.Open()
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer f.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	id := uuid.New().String()
	object := fmt.Sprintf("%s/%s", id, fhdr.Filename)
	fmt.Printf("Object name: %s", object)
	o := client.Bucket("floo-transit").Object(object)
	fmt.Printf("Bucket name: floo-transit")
	wc := o.NewWriter(ctx)
	if _, err = io.Copy(wc, f); err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("Copy done")
	if err := wc.Close(); err != nil {
		fmt.Printf("Writer.Close: %s", err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println("Writer closed")

	attrs, err := o.Attrs(ctx)
	fmt.Println("Attrs obtained")
	if err != nil {
		fmt.Printf("Object(%q).Attrs: %s", object, err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Printf("Media link: %s", attrs.MediaLink)

	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(24 * time.Hour),
	}

	u, err := client.Bucket("floo-transit").SignedURL(object, opts)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	fmt.Println(u)

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
