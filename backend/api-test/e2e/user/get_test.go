//go:build e2e

package user_test

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testapi"
	"github.com/rudradcruze/go-react-template/backend/api-test/fixtures"
)

func cacheKey(id uuid.UUID) string { return "user:user:" + id.String() }

func TestAPI_GetUser_Returns200WithTheUser(t *testing.T) {
	api := testapi.Start(t)
	user := fixtures.SaveUser(t, api.Users, fixtures.WithEmailAddress("ann@example.com"), fixtures.WithDisplayName("Ann"))

	res := api.Get(t, "/api/v1/users/"+user.ID.String())

	require.Equal(t, http.StatusOK, res.StatusCode, string(res.Body))
	got := decode[userBody](t, res)
	assert.Equal(t, user.ID.String(), got.ID)
	assert.Equal(t, "ann@example.com", got.EmailAddress)
	assert.Equal(t, "Ann", got.DisplayName)
	createdAt, err := time.Parse(time.RFC3339Nano, got.CreatedAt)
	require.NoError(t, err)
	assert.True(t, user.CreatedAt.Equal(createdAt), "The API returned %s, and the User has %s.", createdAt, user.CreatedAt)
}

func TestAPI_GetUser_ReturnsAUserThatTheRegistrationMade(t *testing.T) {
	api := testapi.Start(t)
	registered := decode[userBody](t, api.PostJSON(t, "/api/v1/users", `{"emailAddress":"ann@example.com","displayName":"Ann"}`))

	res := api.Get(t, "/api/v1/users/"+registered.ID)

	require.Equal(t, http.StatusOK, res.StatusCode, string(res.Body))
	assert.Equal(t, registered, decode[userBody](t, res))
}

func TestAPI_GetUser_ServesTheSecondReadFromTheCache(t *testing.T) {
	api := testapi.Start(t)
	user := fixtures.SaveUser(t, api.Users)
	path := "/api/v1/users/" + user.ID.String()
	first := api.Get(t, path)
	require.Equal(t, http.StatusOK, first.StatusCode, string(first.Body))

	// Delete the row. Only the cache can still answer the next read.
	_, err := api.DB.ExecContext(t.Context(), "DELETE FROM users WHERE id = $1", user.ID)
	require.NoError(t, err)
	second := api.Get(t, path)

	require.Equal(t, http.StatusOK, second.StatusCode, string(second.Body))
	assert.JSONEq(t, string(first.Body), string(second.Body))
}

func TestAPI_GetUser_KeepsTheUserInTheCacheForFiveMinutes(t *testing.T) {
	api := testapi.Start(t)
	user := fixtures.SaveUser(t, api.Users)

	res := api.Get(t, "/api/v1/users/"+user.ID.String())

	require.Equal(t, http.StatusOK, res.StatusCode, string(res.Body))
	ttl, err := api.Redis.TTL(t.Context(), cacheKey(user.ID)).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, 4*time.Minute+50*time.Second)
	assert.LessOrEqual(t, ttl, 5*time.Minute)
}

func TestAPI_GetUser_ReplacesACacheValueThatDoesNotFitTheID(t *testing.T) {
	other := fixtures.User()
	otherJSON, err := json.Marshal(map[string]any{
		"id": other.ID, "emailAddress": other.EmailAddress, "displayName": other.DisplayName, "createdAt": other.CreatedAt,
	})
	require.NoError(t, err)
	values := map[string]string{
		"empty object":    `{}`,
		"null":            `null`,
		"not JSON":        `not json`,
		"User of another": string(otherJSON),
	}
	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			api := testapi.Start(t)
			user := fixtures.SaveUser(t, api.Users, fixtures.WithDisplayName("Ann"))
			require.NoError(t, api.Redis.Set(t.Context(), cacheKey(user.ID), value, time.Minute).Err())

			res := api.Get(t, "/api/v1/users/"+user.ID.String())

			require.Equal(t, http.StatusOK, res.StatusCode, string(res.Body))
			got := decode[userBody](t, res)
			assert.Equal(t, user.ID.String(), got.ID)
			assert.Equal(t, "Ann", got.DisplayName)
			stored, err := api.Redis.Get(t.Context(), cacheKey(user.ID)).Result()
			require.NoError(t, err)
			assert.Contains(t, stored, user.ID.String(), "The cache still has the value that does not fit.")
		})
	}
}

func TestAPI_GetUser_Returns404AndDoesNotCacheForAnIDThatNoUserHas(t *testing.T) {
	api := testapi.Start(t)
	id := uuid.Must(uuid.NewV7())

	res := api.Get(t, "/api/v1/users/"+id.String())

	requireProblem(t, res, http.StatusNotFound, "does not exist")
	exists, err := api.Redis.Exists(t.Context(), cacheKey(id)).Result()
	require.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestAPI_GetUser_Returns400ForAnIDThatIsNotAUUID(t *testing.T) {
	ids := map[string]string{
		"word":            "abc",
		"number":          "123",
		"too short":       "123e4567-e89b-12d3-a456-42661417400",
		"without dashes":  "123e4567e89b12d3a456426614174000",
		"URN form":        "urn:uuid:123e4567-e89b-12d3-a456-426614174000",
		"with braces":     "%7B123e4567-e89b-12d3-a456-426614174000%7D",
		"not hexadecimal": "123e4567-e89b-12d3-a456-42661417400g",
	}
	api := testapi.Start(t) // The cases change no data, so they share one API.
	for name, id := range ids {
		t.Run(name, func(t *testing.T) {
			res := api.Get(t, "/api/v1/users/"+id)

			requireProblem(t, res, http.StatusBadRequest, "ID")
		})
	}
}

func TestAPI_GetUser_ReadsPostgresWhenRedisFails(t *testing.T) {
	api := testapi.Start(t)
	user := fixtures.SaveUser(t, api.Users, fixtures.WithDisplayName("Ann"))
	api.BreakCache(t)

	res := api.Get(t, "/api/v1/users/"+user.ID.String())

	require.Equal(t, http.StatusOK, res.StatusCode, string(res.Body))
	assert.Equal(t, "Ann", decode[userBody](t, res).DisplayName)
}

func TestAPI_GetUser_Returns500AndHidesTheCauseWhenThePostgresFails(t *testing.T) {
	api := testapi.Start(t)
	user := fixtures.SaveUser(t, api.Users)
	api.BreakDatabase(t)

	res := api.Get(t, "/api/v1/users/"+user.ID.String())

	requireProblem(t, res, http.StatusInternalServerError, "Try again")
	assert.Contains(t, api.Logs(), "The application did not complete the request.")
	assert.Contains(t, api.Logs(), `"route":"/api/v1/users/:id"`)
	assert.Contains(t, api.Logs(), "database is closed")
}
