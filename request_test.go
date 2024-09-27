package gohans

import (
	"context"
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/madflojo/testcerts"
	"github.com/stretchr/testify/assert"
)

func TestNewRequest(t *testing.T) {
	request := NewRequest()

	assert.NotNil(t, request)
	assert.Equal(t, request.Method, "GET")
	assert.Equal(t, request.errorResponse, &Error{})
	assert.Equal(t, request.expectedStatusCode, 200)
	assert.NotNil(t, request.Headers)
}

func TestRequest_SetMethod(t *testing.T) {
	request := NewRequest()
	request.SetMethod("POST")

	assert.Equal(t, request.Method, "POST")
}

func TestRequest_SetURL(t *testing.T) {
	request := NewRequest()
	u, _ := url.Parse("http://example.com")

	request.SetURL(u.String())

	assert.Equal(t, request.URL, "http://example.com")
}

func TestRequest_SetExpectedStatusCode(t *testing.T) {
	request := NewRequest()
	request.SetExpectedStatusCode(201)

	assert.Equal(t, request.expectedStatusCode, 201)
}

func TestRequest_SetErrorResponseBody(t *testing.T) {
	request := NewRequest()
	errorResponse := &struct {
		Err string `json:"err"`
	}{}

	request.SetErrorResponseBody(errorResponse)

	assert.Equal(t, request.errorResponse, errorResponse)
}

func TestRequest_SetWantedResponseBody(t *testing.T) {
	request := NewRequest()
	response := struct {
		Key string `json:"key"`
	}{}

	request.SetWantedResponseBody(&response)

	assert.Equal(t, request.response, &response)
}

func TestRequest_EnableRetries(t *testing.T) {
	request := NewRequest()
	request.EnableRetries(3)

	assert.Equal(t, request.retries, 3)
}

func TestRequest_AddHeader(t *testing.T) {
	request := NewRequest()
	request.AddHeader("key", "value")

	assert.Equal(t, request.Headers["key"], "value")
}

func TestRequest_SetRequestBody(t *testing.T) {
	request := NewRequest()
	request.SetRequestBody("body")

	assert.Equal(t, request.Body, "body")
}

func TestRequest_Send(t *testing.T) {
	ctx := context.Background()
	client := NewClient(ctx)

	t.Run("success", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/test", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Path = "/test"

		var ok struct {
			Status string `json:"status"`
		}

		r := NewRequest().
			SetMethod("GET").
			SetURL(u.String()).
			SetAuthToken("token").
			AddHeader("Content-Type", "application/json").
			SetExpectedStatusCode(200).
			SetWantedResponseBody(&ok)

		body, err := r.Send(ctx, client)

		assert.Nil(t, err)
		assert.Equal(t, `{"status": "ok"}`, string(body))
		assert.Equal(t, "ok", ok.Status)
		assert.Equal(t, &ok, r.GetResponse())

		assert.Equal(t, 200, r.GetStatusCode())
	})

	t.Run("failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "error creating response"}`))
		}))
		defer server.Close()
		u, _ := url.Parse(server.URL)

		var ok struct {
			Status string `json:"status"`
		}
		var e Error

		r := NewRequest().
			SetMethod("GET").
			SetURL(u.String()).
			SetExpectedStatusCode(200).
			SetWantedResponseBody(&ok).
			SetErrorResponseBody(&e)

		body, err := r.Send(ctx, client)

		assert.Error(t, err)
		assert.Equal(t, `{"error": "error creating response"}`, string(body))
		assert.Equal(t, "", ok.Status)
		assert.Equal(t, &ok, r.GetResponse())
		assert.Equal(t, &e, r.GetErrorResponse())
		assert.Equal(t, 500, r.GetStatusCode())
	})

	t.Run("missing url", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/test", r.URL.Path)

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Path = "/test"

		var ok struct {
			Status string `json:"status"`
		}

		r := NewRequest().
			SetMethod("GET").
			SetAuthToken("token").
			SetExpectedStatusCode(200).
			SetWantedResponseBody(&ok)

		body, err := r.Send(ctx, client)

		assert.Error(t, err)
		assert.Equal(t, MissingURLError, err)

		assert.Empty(t, body)
	})

	t.Run("retry", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Retry-Count") == "2" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status": "ok"}`))

				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "error creating response"}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)

		var ok struct {
			Status string `json:"status"`
		}

		r := NewRequest().
			SetMethod("GET").
			SetURL(u.String()).
			SetExpectedStatusCode(200).
			SetWantedResponseBody(&ok).
			EnableRetries(3)

		assert.Equal(t, 3, r.retries)

		body, err := r.Send(ctx, client)

		assert.Nil(t, err)
		assert.Equal(t, `{"status": "ok"}`, string(body))
		assert.Equal(t, "ok", ok.Status)
		assert.Equal(t, &ok, r.GetResponse())
	})

	t.Run("https & tls certs", func(t *testing.T) {
		ctx := context.Background()
		// Generate Certificate Authority
		ca := testcerts.NewCA()
		ca.ToFile("/tmp/ca.crt", "/tmp/ca.key")

		// Create a signed Certificate and Key for "localhost"
		certs, err := ca.NewKeyPair("localhost")
		assert.NoError(t, err)

		// Write certificates to a file
		err = certs.ToFile("/tmp/cert.crt", "/tmp/key.key")
		assert.NoError(t, err)
		tlsConf := tls.Config{
			InsecureSkipVerify: true,
		}
		assert.Nil(t, err)

		client := NewClient(ctx, WithTLSClientConfig(&tlsConf))

		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status": "ok"}`))
		}))
		defer server.Close()

		u, _ := url.Parse(server.URL)
		u.Path = "/test"

		var ok struct {
			Status string `json:"status"`
		}

		r := NewRequest().
			SetMethod("GET").
			SetURL(u.String()).
			SetAuthToken("token").
			SetExpectedStatusCode(200).
			SetWantedResponseBody(&ok)

		body, err := r.Send(ctx, client)

		assert.Nil(t, err)
		assert.Equal(t, `{"status": "ok"}`, string(body))
		assert.Equal(t, "ok", ok.Status)
		assert.Equal(t, &ok, r.GetResponse())

		assert.Equal(t, 200, r.GetStatusCode())
	})

}
