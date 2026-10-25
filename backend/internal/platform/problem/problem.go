// Package problem writes the error responses of the API, as RFC 9457 problem details.
package problem

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Body is an RFC 9457 problem detail.
type Body struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// Write sends a problem detail and stops the handler chain. The detail states the problem,
// and then tells the client what to do.
//
// The title is the HTTP status phrase, for example "Conflict". RFC 9457 asks for it when the type
// is about:blank. It is the one text of the API that is not written in STE.
func Write(c *gin.Context, status int, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, Body{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}

// WriteInternalError sends the response for an unexpected error. It does not show the cause.
func WriteInternalError(c *gin.Context) {
	Write(c, http.StatusInternalServerError, "The server cannot complete the request. Try again in 1 minute.")
}
