//go:build e2e

package platform_test

import (
	"os"
	"testing"

	"github.com/rudradcruze/go-react-template/backend/api-test/e2e/testenv"
)

func TestMain(m *testing.M) {
	// These tests do not need the schema of the application. They use a small test migration.
	testenv.MigrationsDir = "testdata/migrations"
	os.Exit(testenv.Main(m))
}
