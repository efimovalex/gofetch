package gofetch

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

var (
	defaultHeaders = map[string]string{
		"Content-Type": "application/json",
		"Accept":       "application/json",
	}
)

// Error is a struct that represents a response with an error field
type Error struct {
	Error string `json:"error"`
}

// RequestClient is an interface that defines the Do method
type RequestClient interface {
	Do(context.Context, *Request) ([]byte, error)
}

// Request is a struct that represents a request intent
type Request struct {
	Method  string
	URL     url.URL
	Headers map[string]string
	Body    any

	//
	retries int

	// Response and ErrorResponse are used to store the response and error response
	expectedStatusCode int
	response           any
	errorResponse      any
	statusCode         int
}

// NewRequest returns a new Request type with default values
// Method: GET
// Headers: Content-Type: application/json, Accept: application/json
// ErrorResponse: &Error{}
// ExpectedStatusCode: 200
func NewRequest() *Request {
	return &Request{
		Method:             http.MethodGet,
		errorResponse:      &Error{},
		expectedStatusCode: 200,
		Headers:            defaultHeaders,
	}
}

// SetMethod sets the method of the request
func (r *Request) SetMethod(method string) *Request {
	r.Method = method

	return r
}

// SetAuthToken sets the Authorization header with the token
func (r *Request) SetAuthToken(token string) *Request {
	r.Headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

	return r
}

// SetURL sets the URL of the request
func (r *Request) SetURL(url url.URL) *Request {
	r.URL = url

	return r
}

// SetRequestBody sets the body of the request
func (r *Request) SetRequestBody(body interface{}) *Request {
	r.Body = body

	return r
}

// SetWantedResponseBody sets the wanted response body struct
func (r *Request) SetWantedResponseBody(responseBody interface{}) *Request {
	r.response = responseBody

	return r
}

// SetErrorResponseBody sets the error response body struct
func (r *Request) SetErrorResponseBody(errorBody interface{}) *Request {
	r.errorResponse = errorBody

	return r
}

// SetExpectedStatusCode sets the expected status code of the response
func (r *Request) SetExpectedStatusCode(expectedStatusCode int) *Request {
	r.expectedStatusCode = expectedStatusCode

	return r
}

// EnableRetries sets the number of retries for the request
func (r *Request) EnableRetries(retries int) *Request {
	r.retries = retries

	return r
}

// AddHeader adds a header to the request
func (r *Request) AddHeader(key, value string) *Request {
	r.Headers[key] = value

	return r
}

// Do sends the request and returns the response body as a byte slice
// This will retry the request if the number of retries is set
func (r *Request) Do(ctx context.Context, c RequestClient) ([]byte, error) {
	if r.retries > 0 {
		for i := 0; i < r.retries; i++ {
			r.AddHeader("Retry-Count", fmt.Sprint(i))
			body, err := c.Do(ctx, r)
			if err == nil {
				return body, nil
			}
		}
	}

	return c.Do(ctx, r)
}

// GetResponse returns the decoded response body, if successful
func (r *Request) GetResponse() any {
	return r.response
}

// GetErrorResponse returns the decoded error response body, if unsuccessful and the response body could be decoded
func (r *Request) GetErrorResponse() any {
	return r.errorResponse
}

// GetStatusCode returns the status code of the response
func (r *Request) GetStatusCode() int {
	return r.statusCode
}
