// Package config reads the settings of the application from environment variables.
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"reflect"

	"github.com/caarlos0/env/v11"
)

// Config holds all settings of the API.
type Config struct {
	LogLevel slog.Level `env:"LOG_LEVEL" envDefault:"info"`
	APIPort  int        `env:"API_PORT" envDefault:"5653"`
	Docs     Docs
}

// Docs holds the settings of the API docs.
type Docs struct {
	Enabled     bool   `env:"API_DOCS_ENABLED" envDefault:"false"`
	AuthEnabled bool   `env:"API_DOCS_AUTH_ENABLED" envDefault:"true"`
	Username    string `env:"API_DOCS_USERNAME"`
	Password    string `env:"API_DOCS_PASSWORD"`
}

// Load reads the settings from the environment of the process.
func Load() (Config, error) {
	return LoadFrom(env.ToMap(os.Environ()))
}

// LoadFrom reads the settings from the given variables.
// It ignores the environment of the process, also when environment is nil.
func LoadFrom(environment map[string]string) (Config, error) {
	if environment == nil {
		environment = map[string]string{}
	}
	cfg, err := env.ParseAsWithOptions[Config](env.Options{Environment: environment})
	if err != nil {
		return Config{}, nameVariables(err)
	}
	if err := cfg.validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func (c Config) validate() error {
	var errs []error
	if c.APIPort < 1 || c.APIPort > 65535 {
		errs = append(errs, errors.New("API_PORT is not valid. Use a number from 1 to 65535."))
	}
	if c.Docs.Enabled && c.Docs.AuthEnabled && (c.Docs.Username == "" || c.Docs.Password == "") {
		errs = append(errs, errors.New("The authentication of the API docs is on, "+
			"but API_DOCS_USERNAME or API_DOCS_PASSWORD is empty. "+
			"Set both variables. Or set API_DOCS_AUTH_ENABLED to false."))
	}
	return errors.Join(errs...)
}

// nameVariables changes the field names in the parse errors of the env library
// to the names of the environment variables.
func nameVariables(err error) error {
	var aggregate env.AggregateError
	if !errors.As(err, &aggregate) {
		return err
	}
	errs := make([]error, 0, len(aggregate.Errors))
	for _, e := range aggregate.Errors {
		var parseErr env.ParseError
		if errors.As(e, &parseErr) {
			name, _ := variableName(reflect.TypeFor[Config](), parseErr.Name)
			e = fmt.Errorf("%s is not valid. Correct the value. Cause: %w", name, parseErr.Err)
		}
		errs = append(errs, e)
	}
	return errors.Join(errs...)
}

// variableName finds the environment variable of a field in a struct, or in a struct inside it.
// shortcut: the first field with this name wins, so rename a field when 2 nested structs use the same name.
func variableName(t reflect.Type, field string) (string, bool) {
	for f := range t.Fields() {
		if f.Name == field && f.Tag.Get("env") != "" {
			return f.Tag.Get("env"), true
		}
		if f.Type.Kind() == reflect.Struct {
			if name, ok := variableName(f.Type, field); ok {
				return name, true
			}
		}
	}
	return field, false
}
