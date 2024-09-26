package main

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/efimovalex/gohans"
)

type success struct {
	UserId int    `json:"userId"`
	Id     int    `json:"id"`
	Title  string `json:"Title"`
}

type fail struct {
	Error string `json:"status"`
}

func main() {
	logger := slog.Default()
	ctx := context.Background()
	// Create a new client with custom logger
	client := gohans.NewClient(ctx, gohans.WithLogger(logger))

	u := url.URL{Scheme: "https", Host: "jsonplaceholder.typicode.com", Path: "todos/1"}

	var s success
	var f fail

	logger.Info("Example for 200 OK case")
	// Create a new request
	req := gohans.NewRequest().
		SetMethod("GET").
		SetURL(u.String()).
		SetExpectedStatusCode(200).
		SetWantedResponseBody(&s).
		SetErrorResponseBody(&f)

	b, err := req.Send(ctx, client)
	if err != nil {
		logger.Error("Error executing request", "error", err)
		logger.Debug("Response body", "body", string(b))
		logger.Debug("Decoded error", "msg", f.Error)
		return
	}

	logger.Info("Succesful request", "obj", s)

	logger.Info("Example for 404 Not Found case, expecting 200")
	u = url.URL{Scheme: "http", Host: "jsonplaceholder.typicode.com", Path: "todos/0"}
	req = gohans.NewRequest().
		SetMethod(http.MethodGet).
		SetURL(u.String()).
		SetExpectedStatusCode(200).
		SetWantedResponseBody(&s).
		SetErrorResponseBody(&f)

	b, err = req.Send(ctx, client)
	if err != nil {
		logger.Error("Error executing request", "error", err, "body", string(b))
		if req.GetStatusCode() == http.StatusNotFound {
			logger.Error("Entity not found")
		}
		return
	}
}
