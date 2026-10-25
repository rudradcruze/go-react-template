package server

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/problem"
)

// basicAuth checks the credentials of each request in constant time.
// It compares hashes, so the length of the secret does not show in the time.
// When a credential is empty, it rejects all requests.
func basicAuth(username, password string) gin.HandlerFunc {
	rejectAll := username == "" || password == ""
	wantUser := sha256.Sum256([]byte(username))
	wantPass := sha256.Sum256([]byte(password))
	return func(c *gin.Context) {
		user, pass, ok := c.Request.BasicAuth()
		gotUser := sha256.Sum256([]byte(user))
		gotPass := sha256.Sum256([]byte(pass))
		// Both comparisons run before the check, so a wrong username is not faster to reject.
		userMatches := subtle.ConstantTimeCompare(gotUser[:], wantUser[:]) == 1
		passMatches := subtle.ConstantTimeCompare(gotPass[:], wantPass[:]) == 1
		if rejectAll || !ok || !userMatches || !passMatches {
			c.Header("WWW-Authenticate", `Basic realm="API docs", charset="UTF-8"`)
			problem.Write(c, http.StatusUnauthorized,
				"The credentials are missing or not correct. Enter the credentials of the API docs.")
			return
		}
		c.Next()
	}
}
