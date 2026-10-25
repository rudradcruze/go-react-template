package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rudradcruze/go-react-template/backend/api"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
)

// The Scalar script is pinned to one version, and the browser checks its hash.
const docsPage = `<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>API docs</title>
  </head>
  <body>
    <div id="app"></div>
    <script
      src="https://cdn.jsdelivr.net/npm/@scalar/api-reference@1.74.0/dist/browser/standalone.js"
      integrity="sha384-VXnKAz61Xwc3tbw1P+KKrUP7YbcsSl1Ik4is5UnogJyw9XEYZ5Ay/T1i0zPYtSZH"
      crossorigin="anonymous"
    ></script>
    <script>
      Scalar.createApiReference('#app', { url: '/openapi.yaml' })
    </script>
  </body>
</html>
`

// registerDocs adds the docs routes. Without them, Gin sends 404.
func registerDocs(engine *gin.Engine, docs config.Docs) {
	if !docs.Enabled {
		return
	}
	group := engine.Group("/")
	if docs.AuthEnabled {
		group.Use(basicAuth(docs.Username, docs.Password))
	}
	group.GET("/docs", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(docsPage))
	})
	group.GET("/openapi.yaml", func(c *gin.Context) {
		c.Data(http.StatusOK, "application/yaml", api.Spec)
	})
}
