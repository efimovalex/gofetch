package gohans

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

var (
	DecodeError               = errors.New("error decoding response")
	UnexpectedStatusCodeError = errors.New("unexpected status code")
	MissingURLError           = errors.New("URL is missing")
)

type RequestOption func(*Client)

type Client struct {
	logger *slog.Logger

	httpClient *http.Client
}

func NewClient(ctx context.Context, opts ...RequestOption) *Client {
	c := &Client{
		httpClient: &http.Client{},
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.logger == nil {
		c.logger = slog.Default()
	}

	return c
}

// WithTLSClientConfig sets the TLSClientConfig on the http client
func WithTLSClientConfig(tlsConfig *tls.Config) RequestOption {
	return func(c *Client) {
		t := &http.Transport{
			TLSClientConfig: tlsConfig,
		}

		c.httpClient.Transport = t
	}
}

// WithTimeout sets a request timeout on the client
func WithTimeout(td time.Duration) RequestOption {
	return func(c *Client) {
		c.httpClient.Timeout = td
	}
}

// WithHTTPClient sets the http client on the client
// Warning: This will override any other transport settings set prior to this
func WithHTTPClient(client *http.Client) RequestOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithLogger(logger *slog.Logger) RequestOption {
	return func(c *Client) {
		c.logger = logger
	}
}

// Do sends a request adds the decoded values from the response to the request object
// Returns the response body as a byte slice for debugging or further processing
// If the response status code is not the expected status code, we try to decode the response body into the error response object
// If the response body cannot be decoded into the error response object, we return an error

func (c *Client) Do(ctx context.Context, r *Request) ([]byte, error) {
	var br bytes.Buffer

	if r.URL == "" {
		c.logger.Error("URL is not set")

		return nil, MissingURLError
	}

	url, err := url.Parse(r.URL)
	if err != nil {
		c.logger.Error("Malformed URL", "url", r.URL)

		return nil, err
	}

	if r.Body != nil {
		err := json.NewEncoder(&br).Encode(r.Body)
		if err != nil {
			c.logger.Error("error encoding request body", "error", err)
			return nil, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, url.String(), &br)
	if err != nil {
		c.logger.Error("error creating request", "error", err)
		return nil, err
	}

	for k, v := range r.Headers {
		req.Header.Add(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("error sending request", "error", err)
		return nil, err
	}
	defer resp.Body.Close()

	var buf bytes.Buffer
	tee := io.TeeReader(resp.Body, &buf)

	r.statusCode = resp.StatusCode

	if resp.StatusCode != r.expectedStatusCode {
		c.logger.Error("unexpected status code", "expected", r.expectedStatusCode, "actual", resp.StatusCode)
		err = json.NewDecoder(tee).Decode(&r.errorResponse)
		if err != nil {
			c.logger.Error("error decoding error response", "error", err)
			return buf.Bytes(), err
		}

		return buf.Bytes(), UnexpectedStatusCodeError
	}

	err = json.NewDecoder(tee).Decode(&r.response)
	if err != nil {
		c.logger.Error("error decoding response", "error", err)

		return buf.Bytes(), err
	}

	return buf.Bytes(), nil
}
