package main

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/labstack/echo/v4"
	"google.golang.org/api/iterator"
)

func DownloadFile(c echo.Context) error {
	id := c.Param("uuid")
	ctx := context.Background()
	fireclnt, err := firestore.NewClient(ctx, "floo-network")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer fireclnt.Close()

	iter := fireclnt.Collection("files").Where("uuid", "==", id).Documents(ctx)
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		signed_url := fmt.Sprint(doc.Data()["signed_url"])
		return c.Redirect(http.StatusMovedPermanently, signed_url)
	}
	return c.NoContent(http.StatusNotFound)
}
