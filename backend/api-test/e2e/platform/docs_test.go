//go:build e2e

package platform_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pb33f/libopenapi"
	schemavalidation "github.com/pb33f/libopenapi-validator/schema_validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/server"
)

var docsRoutes = []string{"/docs", "/openapi.yaml"}

// startServer starts the real Gin engine on a test port.
func startServer(t *testing.T, docs config.Docs) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(server.New(config.Config{Docs: docs}, slog.New(slog.NewTextHandler(io.Discard, nil))))
	t.Cleanup(srv.Close)
	return srv
}

// response is the part of an HTTP response that the tests check.
type response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

type basicCredentials struct {
	username string
	password string
}

// get sends a GET request. It sets Basic auth only when the test gives credentials.
func get(t *testing.T, srv *httptest.Server, path string, credentials ...basicCredentials) response {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL+path, http.NoBody)
	require.NoError(t, err)
	for _, c := range credentials {
		req.SetBasicAuth(c.username, c.password)
	}
	res, err := srv.Client().Do(req)
	require.NoError(t, err)
	defer func() { _ = res.Body.Close() }()
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	return response{StatusCode: res.StatusCode, Header: res.Header, Body: body}
}

func TestServer_New_Sends404ForTheDocsRoutesWhenDocsAreOff(t *testing.T) {
	srv := startServer(t, config.Docs{Enabled: false, AuthEnabled: true, Username: "reader", Password: "secret"})

	for _, path := range docsRoutes {
		t.Run(path, func(t *testing.T) {
			res := get(t, srv, path, basicCredentials{"reader", "secret"})

			assert.Equal(t, http.StatusNotFound, res.StatusCode)
			assert.Equal(t, "application/problem+json", res.Header.Get("Content-Type"))
		})
	}
}

func TestServer_New_ServesTheDocsWithoutCredentialsWhenAuthIsOff(t *testing.T) {
	srv := startServer(t, config.Docs{Enabled: true, AuthEnabled: false})

	page := get(t, srv, "/docs")
	spec := get(t, srv, "/openapi.yaml")

	assert.Equal(t, http.StatusOK, page.StatusCode)
	assert.Contains(t, page.Header.Get("Content-Type"), "text/html")
	assert.Equal(t, http.StatusOK, spec.StatusCode)
	assert.Contains(t, spec.Header.Get("Content-Type"), "yaml")
}

func TestServer_New_RequiresCorrectCredentialsForTheDocsWhenAuthIsOn(t *testing.T) {
	srv := startServer(t, config.Docs{Enabled: true, AuthEnabled: true, Username: "reader", Password: "secret"})
	tests := []struct {
		name        string
		credentials []basicCredentials
		want        int
	}{
		{name: "no credentials", want: http.StatusUnauthorized},
		{name: "wrong password", credentials: []basicCredentials{{"reader", "wrong"}}, want: http.StatusUnauthorized},
		{name: "wrong username", credentials: []basicCredentials{{"other", "secret"}}, want: http.StatusUnauthorized},
		{name: "correct credentials", credentials: []basicCredentials{{"reader", "secret"}}, want: http.StatusOK},
	}
	for _, path := range docsRoutes {
		for _, tt := range tests {
			t.Run(path+"/"+tt.name, func(t *testing.T) {
				res := get(t, srv, path, tt.credentials...)

				assert.Equal(t, tt.want, res.StatusCode)
				if tt.want == http.StatusUnauthorized {
					assert.Contains(t, res.Header.Get("WWW-Authenticate"), "Basic")
					assert.Equal(t, "application/problem+json", res.Header.Get("Content-Type"))
				}
			})
		}
	}
}

// config.Load rejects empty credentials. This test checks the second line of defence.
func TestServer_New_RejectsAllDocsRequestsWhenAuthHasEmptyCredentials(t *testing.T) {
	srv := startServer(t, config.Docs{Enabled: true, AuthEnabled: true})

	for _, path := range docsRoutes {
		t.Run(path, func(t *testing.T) {
			withoutCredentials := get(t, srv, path)
			withEmptyCredentials := get(t, srv, path, basicCredentials{"", ""})

			assert.Equal(t, http.StatusUnauthorized, withoutCredentials.StatusCode)
			assert.Equal(t, http.StatusUnauthorized, withEmptyCredentials.StatusCode)
		})
	}
}

func TestServer_New_ServesAValidOpenAPI32Spec(t *testing.T) {
	srv := startServer(t, config.Docs{Enabled: true, AuthEnabled: false})
	res := get(t, srv, "/openapi.yaml")

	doc, err := libopenapi.NewDocument(res.Body)
	require.NoError(t, err)
	valid, validationErrs := schemavalidation.ValidateOpenAPIDocument(doc)

	assert.Equal(t, "3.2.0", doc.GetSpecInfo().Version)
	assert.True(t, valid, "the spec is not valid: %v", validationErrs)
}
