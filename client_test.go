package gofetch

import (
	"crypto/tls"
	"log/slog"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"context"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default()
	client := NewClient(ctx)

	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, client.logger, logger)
}

func TestWithTLSClientConfig(t *testing.T) {
	ctx := context.Background()

	tlsc := &tls.Config{}

	client := NewClient(ctx, WithTLSClientConfig(tlsc))
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient.Transport)
}

func TestWithHTTPClient(t *testing.T) {
	ctx := context.Background()

	httpClient := &http.Client{}

	client := NewClient(ctx, WithHTTPClient(httpClient))
	assert.NotNil(t, client)
	assert.Equal(t, client.httpClient, httpClient)
}

func TestWithLogger(t *testing.T) {
	ctx := context.Background()
	logger := slog.Default().With("t", "test")

	client := NewClient(ctx, WithLogger(logger))
	assert.NotNil(t, client)
	assert.Equal(t, client.logger, logger)
}

func TestDo(t *testing.T) {
	ctx := context.Background()

	client := NewClient(ctx)
	assert.NotNil(t, client)
	assert.NotNil(t, client.httpClient)

	t.Run("http server success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		assert.Nil(t, err)

		var ok struct {
			Status string `json:"status"`
		}

		body, err := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(*u).
			SetExpectedStatusCode(http.StatusOK).
			SetErrorResponseBody(&Error{}).
			SetWantedResponseBody(&ok).
			Do(ctx, client)

		assert.NotNil(t, body)
		assert.Nil(t, err)
		assert.Equal(t, `{"status": "ok"}`, string(body))
		assert.Equal(t, "ok", ok.Status)
	})

	t.Run("http server failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "error creating response"}`))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		assert.Nil(t, err)

		var ok struct {
			Status string `json:"status"`
		}
		var e Error

		r := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(*u).
			SetExpectedStatusCode(http.StatusOK).
			SetErrorResponseBody(&e).
			SetWantedResponseBody(&ok)

		body, err := r.Do(ctx, client)

		assert.NotNil(t, body)
		assert.Equal(t, UnexpectedStatusCodeError, err)
		assert.Equal(t, `{"error": "error creating response"}`, string(body))
		assert.Equal(t, "", ok.Status)

		assert.Equal(t, "error creating response", e.Error)
	})

	t.Run("http server failure, map decode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "error creating response"}`))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		assert.Nil(t, err)

		var ok struct {
			Status string `json:"status"`
		}
		var e map[string]interface{}

		r := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(*u).
			SetExpectedStatusCode(http.StatusOK).
			SetErrorResponseBody(&e).
			SetWantedResponseBody(&ok)

		body, err := r.Do(ctx, client)

		assert.NotNil(t, body)
		assert.Equal(t, UnexpectedStatusCodeError, err)
		assert.Equal(t, `{"error": "error creating response"}`, string(body))
		assert.Equal(t, "", ok.Status)

		assert.Equal(t, "error creating response", e["error"].(string))
	})

	t.Run("error decoding  error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{]`))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		assert.Nil(t, err)

		var ok struct {
			Status string `json:"status"`
		}
		var e map[string]interface{}

		r := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(*u).
			SetExpectedStatusCode(http.StatusOK).
			SetErrorResponseBody(&e).
			SetWantedResponseBody(&ok)

		body, err := r.Do(ctx, client)

		assert.NotNil(t, body)
		assert.EqualError(t, err, "invalid character ']' looking for beginning of object key string")
		assert.Equal(t, `{]`, string(body))
		assert.Equal(t, "", ok.Status)
	})

	t.Run("error decoding success response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{]`))
		}))
		defer server.Close()

		u, err := url.Parse(server.URL)
		assert.Nil(t, err)

		var ok struct {
			Status string `json:"status"`
		}
		var e map[string]interface{}

		r := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(*u).
			SetExpectedStatusCode(http.StatusOK).
			SetErrorResponseBody(&e).
			SetWantedResponseBody(&ok)

		body, err := r.Do(ctx, client)

		assert.NotNil(t, body)
		assert.EqualError(t, err, "invalid character ']' looking for beginning of object key string")
		assert.Equal(t, `{]`, string(body))
		assert.Equal(t, "", ok.Status)
	})

	t.Run("error encoding request", func(t *testing.T) {
		r := NewRequest().
			SetMethod(http.MethodGet).
			SetRequestBody(struct{ A float64 }{A: math.Inf(0)})

		body, err := r.Do(ctx, client)
		assert.Nil(t, body)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "json: unsupported value")
	})

	t.Run("error doing request", func(t *testing.T) {
		r := NewRequest().
			SetMethod(http.MethodGet).
			SetURL(url.URL{Scheme: "w", Host: "localhost:0"})

		body, err := r.Do(ctx, client)
		assert.Nil(t, body)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "Get \"w://localhost:0\": unsupported protocol scheme \"w\"")
	})

	t.Run("error invalid method", func(t *testing.T) {
		r := NewRequest().
			SetMethod("💀")

		body, err := r.Do(ctx, client)
		assert.Nil(t, body)
		assert.Error(t, err)
		assert.Equal(t, err.Error(), "net/http: invalid method \"💀\"")
	})

}
