package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// problem is an RFC 9457 problem detail.
type problem struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Status int    `json:"status"`
	Detail string `json:"detail"`
}

// writeProblem sends a problem detail and stops the handler chain.
func writeProblem(c *gin.Context, status int, detail string) {
	c.Header("Content-Type", "application/problem+json")
	c.AbortWithStatusJSON(status, problem{
		Type:   "about:blank",
		Title:  http.StatusText(status),
		Status: status,
		Detail: detail,
	})
}
