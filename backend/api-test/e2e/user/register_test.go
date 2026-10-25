//go:build e2e

package user_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testapi"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/problem"
)

type userBody struct {
	ID           string `json:"id"`
	EmailAddress string `json:"emailAddress"`
	DisplayName  string `json:"displayName"`
	CreatedAt    string `json:"createdAt"`
}

func decode[T any](t *testing.T, res testapi.Response) T {
	t.Helper()
	var v T
	require.NoError(t, json.Unmarshal(res.Body, &v), string(res.Body))
	return v
}

func countUsers(t *testing.T, api *testapi.API) int {
	t.Helper()
	var n int
	require.NoError(t, api.DB.QueryRowContext(t.Context(), "SELECT count(*) FROM users").Scan(&n))
	return n
}

func requireProblem(t *testing.T, res testapi.Response, status int, detailPart string) {
	t.Helper()
	require.Equal(t, status, res.StatusCode, string(res.Body))
	assert.Equal(t, "application/problem+json", res.Header.Get("Content-Type"))
	problem := decode[problem.Body](t, res)
	assert.Equal(t, status, problem.Status)
	assert.Equal(t, http.StatusText(status), problem.Title)
	assert.Contains(t, problem.Detail, detailPart)
}

func TestAPI_RegisterUser_Returns201AndSavesTheUser(t *testing.T) {
	api := testapi.Start(t)

	res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"  Ann@Example.COM ","displayName":" Ann "}`)

	require.Equal(t, http.StatusCreated, res.StatusCode, string(res.Body))
	assert.Equal(t, "application/json; charset=utf-8", res.Header.Get("Content-Type"))
	got := decode[userBody](t, res)
	id, err := uuid.Parse(got.ID)
	require.NoError(t, err)
	assert.Equal(t, uuid.Version(7), id.Version())
	assert.Equal(t, "ann@example.com", got.EmailAddress)
	assert.Equal(t, "Ann", got.DisplayName)
	createdAt, err := time.Parse(time.RFC3339Nano, got.CreatedAt)
	require.NoError(t, err)
	assert.WithinDuration(t, time.Now(), createdAt, time.Minute)

	var email, name string
	var stored time.Time
	err = api.DB.QueryRowContext(t.Context(),
		"SELECT email_address, display_name, created_at FROM users WHERE id = $1", id).Scan(&email, &name, &stored)
	require.NoError(t, err)
	assert.Equal(t, "ann@example.com", email)
	assert.Equal(t, "Ann", name)
	assert.True(t, createdAt.Equal(stored), "The API returned %s, and the database has %s.", createdAt, stored)
}

func TestAPI_RegisterUser_Returns409ForAnEmailAddressThatIsRegistered(t *testing.T) {
	tests := map[string]string{
		"same spelling":     "ann@example.com",
		"other letter case": "ANN@Example.COM",
		"spaces around":     "  ann@example.com  ",
	}
	for name, second := range tests {
		t.Run(name, func(t *testing.T) {
			api := testapi.Start(t)
			first := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`)
			require.Equal(t, http.StatusCreated, first.StatusCode, string(first.Body))

			res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"`+second+`","displayName":"Other"}`)

			requireProblem(t, res, http.StatusConflict, "already registered")
			assert.Equal(t, 1, countUsers(t, api))
		})
	}
}

func TestAPI_RegisterUser_Returns422WhenARuleRejectsTheInput(t *testing.T) {
	tests := map[string]struct {
		body       string
		detailPart string
	}{
		"email address is not valid": {`{"emailAddress":"ann","displayName":"Ann"}`, "email address is not valid"},
		"email address is missing":   {`{"displayName":"Ann"}`, "email address is not valid"},
		"display name is empty":      {`{"emailAddress":"ann@example.com","displayName":"  "}`, "display name is empty"},
		"display name is missing":    {`{"emailAddress":"ann@example.com"}`, "display name is empty"},
		"display name is too long": {
			`{"emailAddress":"ann@example.com","displayName":"` + strings.Repeat("a", 101) + `"}`,
			"display name is too long",
		},
		"display name has a NUL character": {
			`{"emailAddress":"ann@example.com","displayName":"a\u0000b"}`,
			"character that is not allowed",
		},
		"display name has a new line": {
			`{"emailAddress":"ann@example.com","displayName":"a\nb"}`,
			"character that is not allowed",
		},
		"display name is only a zero-width space": {
			`{"emailAddress":"ann@example.com","displayName":"\u200b"}`,
			"display name is empty",
		},
		"email address has other characters than ASCII": {
			`{"emailAddress":"annä@example.com","displayName":"Ann"}`,
			"email address is not valid",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			api := testapi.Start(t)

			res := api.PostJSON(t, "/api/v1/users", tt.body)

			requireProblem(t, res, http.StatusUnprocessableEntity, tt.detailPart)
			assert.Equal(t, 0, countUsers(t, api))
		})
	}
}

func TestAPI_RegisterUser_Returns400ForABodyThatTheAPICannotParse(t *testing.T) {
	tests := map[string]string{
		"empty body":          ``,
		"broken JSON":         `{"emailAddress":`,
		"array":               `[]`,
		"text":                `"ann@example.com"`,
		"field of wrong type": `{"emailAddress":5,"displayName":"Ann"}`,
		"data after the JSON": `{"emailAddress":"ann@example.com","displayName":"Ann"} {"x":1}`,
		"text after the JSON": `{"emailAddress":"ann@example.com","displayName":"Ann"} garbage`,
		"unknown field":       `{"emailAddress":"ann@example.com","displayName":"Ann","nickname":"A"}`,
	}
	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			api := testapi.Start(t)

			res := api.PostJSON(t, "/api/v1/users", body)

			requireProblem(t, res, http.StatusBadRequest, "not valid JSON")
			assert.Equal(t, 0, countUsers(t, api))
		})
	}
}

func TestAPI_RegisterUser_Returns415ForAContentTypeThatIsNotJSON(t *testing.T) {
	body := []byte(`{"emailAddress":"ann@example.com","displayName":"Ann"}`)
	for _, contentType := range []string{"", "text/plain", "application/x-www-form-urlencoded", "application/xml"} {
		t.Run("content type "+contentType, func(t *testing.T) {
			api := testapi.Start(t)

			res := api.Do(t, http.MethodPost, "/api/v1/users", contentType, body)

			requireProblem(t, res, http.StatusUnsupportedMediaType, "application/json")
			assert.Equal(t, 0, countUsers(t, api))
		})
	}
}

func TestAPI_RegisterUser_AcceptsAJSONContentTypeWithACharset(t *testing.T) {
	api := testapi.Start(t)

	res := api.Do(t, http.MethodPost, "/api/v1/users", "application/json; charset=utf-8",
		[]byte(`{"emailAddress":"ann@example.com","displayName":"Ann"}`))

	require.Equal(t, http.StatusCreated, res.StatusCode, string(res.Body))
}

func TestAPI_RegisterUser_Returns413ForABodyLargerThan1MiB(t *testing.T) {
	api := testapi.Start(t)
	body := `{"emailAddress":"ann@example.com","displayName":"` + strings.Repeat("a", 2<<20) + `"}`

	res := api.PostJSON(t, "/api/v1/users", body)

	requireProblem(t, res, http.StatusRequestEntityTooLarge, "1 MiB")
	assert.Equal(t, 0, countUsers(t, api))
}

func TestAPI_RegisterUser_Returns500AndHidesTheCauseWhenTheDatabaseFails(t *testing.T) {
	api := testapi.Start(t)
	api.BreakDatabase(t)

	res := api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`)

	requireProblem(t, res, http.StatusInternalServerError, "Try again")
	assert.NotContains(t, strings.ToLower(string(res.Body)), "sql")
	assert.NotContains(t, strings.ToLower(string(res.Body)), "closed")
}

func TestUsersTable_RejectsARowThatBreaksARuleOfTheDomain(t *testing.T) {
	tests := map[string]struct {
		email string
		name  string
	}{
		"email address with uppercase letters": {"Ann@example.com", "Ann"},
		"email address that is empty":          {"", "Ann"},
		"email address over 254 characters":    {strings.Repeat("a", 250) + "@example.com", "Ann"},
		"display name that is empty":           {"ann@example.com", ""},
		"display name over 100 characters":     {"ann@example.com", strings.Repeat("a", 101)},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			api := testapi.Start(t)

			_, err := api.DB.ExecContext(t.Context(),
				"INSERT INTO users (id, email_address, display_name, created_at) VALUES ($1, $2, $3, now())",
				uuid.New(), tt.email, tt.name)

			require.Error(t, err)
		})
	}
}

func TestUsersTable_RejectsTwoRowsWithTheSameEmailAddress(t *testing.T) {
	api := testapi.Start(t)
	insert := func() error {
		_, err := api.DB.ExecContext(t.Context(),
			"INSERT INTO users (id, email_address, display_name, created_at) VALUES ($1, 'ann@example.com', 'Ann', now())",
			uuid.New())
		return err
	}
	require.NoError(t, insert())

	require.Error(t, insert())
}
