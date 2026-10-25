package platform_test

import (
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
)

func TestConfig_LoadFrom_UsesTheDefaultsWhenTheEnvironmentIsEmpty(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(map[string]string{})

	require.NoError(t, err)
	assert.Equal(t, 5653, cfg.APIPort)
	assert.Equal(t, slog.LevelInfo, cfg.LogLevel)
	assert.False(t, cfg.Docs.Enabled)
	assert.True(t, cfg.Docs.AuthEnabled)
	assert.Empty(t, cfg.Docs.Username)
	assert.Empty(t, cfg.Docs.Password)
}

func TestConfig_LoadFrom_ReadsTheGivenValues(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(map[string]string{
		"LOG_LEVEL":             "debug",
		"API_PORT":              "6000",
		"API_DOCS_ENABLED":      "true",
		"API_DOCS_AUTH_ENABLED": "true",
		"API_DOCS_USERNAME":     "reader",
		"API_DOCS_PASSWORD":     "secret",
	})

	require.NoError(t, err)
	assert.Equal(t, 6000, cfg.APIPort)
	assert.Equal(t, slog.LevelDebug, cfg.LogLevel)
	assert.True(t, cfg.Docs.Enabled)
	assert.True(t, cfg.Docs.AuthEnabled)
	assert.Equal(t, "reader", cfg.Docs.Username)
	assert.Equal(t, "secret", cfg.Docs.Password)
}

func TestConfig_LoadFrom_RejectsDocsAuthWithoutCredentials(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		env  map[string]string
	}{
		{
			name: "no username and no password",
			env:  map[string]string{"API_DOCS_ENABLED": "true"},
		},
		{
			name: "no username",
			env:  map[string]string{"API_DOCS_ENABLED": "true", "API_DOCS_PASSWORD": "secret"},
		},
		{
			name: "no password",
			env:  map[string]string{"API_DOCS_ENABLED": "true", "API_DOCS_USERNAME": "reader"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.LoadFrom(tt.env)

			require.ErrorContains(t, err, "API_DOCS_USERNAME")
			require.ErrorContains(t, err, "API_DOCS_PASSWORD")
		})
	}
}

func TestConfig_LoadFrom_AcceptsDocsWithoutAuthAndWithoutCredentials(t *testing.T) {
	t.Parallel()

	cfg, err := config.LoadFrom(map[string]string{
		"API_DOCS_ENABLED":      "true",
		"API_DOCS_AUTH_ENABLED": "false",
	})

	require.NoError(t, err)
	assert.True(t, cfg.Docs.Enabled)
	assert.False(t, cfg.Docs.AuthEnabled)
}

func TestConfig_LoadFrom_RejectsInvalidValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		env      map[string]string
		variable string
	}{
		{name: "port is not a number", env: map[string]string{"API_PORT": "abc"}, variable: "API_PORT"},
		{name: "port is zero", env: map[string]string{"API_PORT": "0"}, variable: "API_PORT"},
		{name: "port is too high", env: map[string]string{"API_PORT": "70000"}, variable: "API_PORT"},
		{name: "log level is unknown", env: map[string]string{"LOG_LEVEL": "loud"}, variable: "LOG_LEVEL"},
		{name: "docs flag is not a boolean", env: map[string]string{"API_DOCS_ENABLED": "maybe"}, variable: "API_DOCS_ENABLED"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := config.LoadFrom(tt.env)

			require.ErrorContains(t, err, tt.variable)
		})
	}
}

// A person copies .env.example to .env, so the file must give a valid config.
func TestConfig_LoadFrom_AcceptsTheEnvExampleFile(t *testing.T) {
	t.Parallel()
	content, err := os.ReadFile("../../../../.env.example")
	require.NoError(t, err)
	environment := map[string]string{}
	for line := range strings.Lines(string(content)) {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, _ := strings.Cut(line, "=")
		environment[key] = value
	}

	cfg, err := config.LoadFrom(environment)

	require.NoError(t, err)
	assert.Equal(t, 5653, cfg.APIPort)
	assert.False(t, cfg.Docs.Enabled)
}

// The next 2 tests cannot call t.Parallel, because t.Setenv changes the environment of the process.

func TestConfig_Load_ReadsTheEnvironmentOfTheProcess(t *testing.T) {
	t.Setenv("API_PORT", "6001")

	cfg, err := config.Load()

	require.NoError(t, err)
	assert.Equal(t, 6001, cfg.APIPort)
}

func TestConfig_LoadFrom_IgnoresTheEnvironmentOfTheProcess(t *testing.T) {
	t.Setenv("API_PORT", "6002")

	cfg, err := config.LoadFrom(nil)

	require.NoError(t, err)
	assert.Equal(t, 5653, cfg.APIPort)
}
