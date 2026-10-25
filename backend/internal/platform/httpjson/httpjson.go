// Package httpjson reads the JSON body of a request. All handlers use it, so all endpoints
// apply the same rules to a body.
package httpjson

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/problem"
)

// MaxBodyBytes is the maximum size of a request body: 1 MiB. The OpenAPI spec states the same limit.
const MaxBodyBytes = 1 << 20

// Decode reads the JSON body of the request into dst. When the request is not valid, Decode writes
// the problem detail and returns false. The handler must then return.
//
// The problem is 415 for a content type that is not application/json, 413 for a body of more than 1 MiB,
// and 400 for a body that is not valid JSON, has a field that dst does not have, or has data after the JSON value.
func Decode(c *gin.Context, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		problem.Write(c, http.StatusUnsupportedMediaType,
			"The content type is not supported. Send the request with the content type application/json.")
		return false
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dst); err != nil {
		writeBodyError(c, err)
		return false
	}
	// The body must have only one JSON value. The next token must be the end of the body.
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		writeBodyError(c, err)
		return false
	}
	return true
}

func writeBodyError(c *gin.Context, err error) {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		problem.Write(c, http.StatusRequestEntityTooLarge,
			fmt.Sprintf("The request body is larger than %d MiB. Send a smaller body.", MaxBodyBytes>>20))
		return
	}
	problem.Write(c, http.StatusBadRequest,
		"The request body is not valid JSON, or it has a field that the API does not know. "+
			"Send one JSON object with only the fields of the API docs.")
}
