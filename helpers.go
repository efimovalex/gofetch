package gohans

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"net/http"
)

// decodeResponse decodes the response body into the result interface
func decodeResponse(resp *http.Response, result interface{}) error {
	switch resp.Header.Get("Content-Type") {
	case "application/json":
		return json.NewDecoder(resp.Body).Decode(&result)
	case "application/xml":
		return xml.NewDecoder(resp.Body).Decode(&result)
	default:
		return errors.New("invalid content type")
	}
}
