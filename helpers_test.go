package gofetch

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_decodeResponse(t *testing.T) {

	t.Run("json", func(t *testing.T) {

		body := `{"key": "value"}`
		resp := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(body)),
		}

		result := map[string]interface{}{}
		err := decodeResponse(resp, &result)
		assert.NoError(t, err)
		assert.Equal(t, map[string]interface{}{"key": "value"}, result)
	})

	t.Run("xml", func(t *testing.T) {
		body := `<struct><key>value</key></struct>`
		resp := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/xml"},
			},
			Body: io.NopCloser(strings.NewReader(body)),
		}

		result := struct {
			Key string `xml:"key"`
		}{}
		err := decodeResponse(resp, &result)
		assert.NoError(t, err)
		assert.Equal(t, "value", result.Key)
	})

	t.Run("invalid content type", func(t *testing.T) {
		resp := &http.Response{
			Header: http.Header{
				"Content-Type": []string{"application/invalid"},
			},
		}

		err := decodeResponse(resp, nil)
		assert.Error(t, err)
		assert.Equal(t, "invalid content type", err.Error())
	})
}
