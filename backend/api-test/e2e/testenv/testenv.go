//go:build e2e

// Package testenv gives the E2E tests real Postgres and Redis in containers.
//
// The containers start one time for each test package. Each test gets its own database,
// a copy of a template database. Begin drops the database when the test ends.
package testenv

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib" // The "pgx" driver of database/sql.
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"

	"github.com/rudradcruze/go-react-template/backend/internal/platform/config"
	"github.com/rudradcruze/go-react-template/backend/internal/platform/database"
)

const (
	postgresImage = "postgres:18-alpine" // Use the same image as compose.yaml.
	redisImage    = "redis:7-alpine"     // Use the same image as compose.yaml.
	templateDB    = "app_template"
	maintenanceDB = "postgres"
	dbUser        = "app"
	dbPassword    = "test"
)

// Resources are the connection settings that one test uses.
type Resources struct {
	Postgres config.Postgres
	Redis    config.Redis
}

type infra struct {
	resources Resources // The Postgres settings point to the template database.
	admin     *sql.DB   // Connected to the maintenance database.
}

var current *infra

// MigrationsDir is the folder of the migrations that run on the template database.
// A test package can set it in TestMain before it calls Main. When it is empty,
// Main uses backend/migrations.
var MigrationsDir string

// Main starts the containers, runs the tests and stops the containers.
// Call it from TestMain: os.Exit(testenv.Main(m)).
func Main(m *testing.M) int {
	ctx := context.Background()
	in, stop, err := start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "testenv: The containers did not start. Cause:", err)
		return 1
	}
	defer stop()
	current = in
	return m.Run()
}

func start(ctx context.Context) (in *infra, stop func(), err error) {
	var containers []testcontainers.Container
	stopContainers := func() {
		for _, c := range containers {
			_ = testcontainers.TerminateContainer(c)
		}
	}
	defer func() {
		if err != nil {
			stopContainers()
		}
	}()

	pg, err := tcpostgres.Run(ctx, postgresImage,
		tcpostgres.WithDatabase(templateDB),
		tcpostgres.WithUsername(dbUser),
		tcpostgres.WithPassword(dbPassword),
		tcpostgres.BasicWaitStrategies(),
	)
	if pg != nil {
		containers = append(containers, pg)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("The Postgres container did not start. Cause: %w", err)
	}
	rd, err := tcredis.Run(ctx, redisImage)
	if rd != nil {
		containers = append(containers, rd)
	}
	if err != nil {
		return nil, nil, fmt.Errorf("The Redis container did not start. Cause: %w", err)
	}

	pgHost, pgPort, err := endpoint(ctx, pg, "5432/tcp")
	if err != nil {
		return nil, nil, err
	}
	rdHost, rdPort, err := endpoint(ctx, rd, "6379/tcp")
	if err != nil {
		return nil, nil, err
	}
	resources := Resources{
		Postgres: config.Postgres{Host: pgHost, Port: pgPort, User: dbUser, Password: dbPassword, DB: templateDB, SSLMode: "disable"},
		Redis:    config.Redis{Host: rdHost, Port: rdPort, CacheDB: 0, QueueDB: 1},
	}

	if err := migrateTemplate(ctx, resources.Postgres); err != nil {
		return nil, nil, err
	}

	maintenance := resources.Postgres
	maintenance.DB = maintenanceDB
	admin, err := sql.Open("pgx", maintenance.DSN())
	if err != nil {
		return nil, nil, fmt.Errorf("The maintenance database did not open. Cause: %w", err)
	}
	stop = func() {
		_ = admin.Close()
		stopContainers()
	}
	return &infra{resources: resources, admin: admin}, stop, nil
}

// migrateTemplate runs the migrations on the template database, and then closes all connections to it.
// Postgres cannot copy a database that has open connections.
func migrateTemplate(ctx context.Context, template config.Postgres) error {
	db, err := sql.Open("pgx", template.DSN())
	if err != nil {
		return fmt.Errorf("The template database did not open. Cause: %w", err)
	}
	defer func() { _ = db.Close() }()
	dir := MigrationsDir
	if dir == "" {
		_, thisFile, _, _ := runtime.Caller(0)
		dir = filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "migrations")
	}
	return database.MigrateUp(ctx, db, dir)
}

func endpoint(ctx context.Context, c testcontainers.Container, port string) (string, int, error) {
	host, err := c.Host(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("The host of a container is not available. Cause: %w", err)
	}
	mapped, err := c.MappedPort(ctx, port)
	if err != nil {
		return "", 0, fmt.Errorf("The port %s of a container is not available. Cause: %w", port, err)
	}
	return host, int(mapped.Num()), nil
}

// Begin gives the test a new database, a copy of the migrated template database,
// and empties Redis. It drops the database when the test ends, also when the test fails.
//
// Redis has one set of data for all tests, so a test that calls Begin must not call t.Parallel.
func Begin(t *testing.T) Resources {
	t.Helper()
	return begin(t, true)
}

// BeginEmpty is like Begin, but the new database has no tables.
// Use it to test the migrations, because the template database has applied migrations.
func BeginEmpty(t *testing.T) Resources {
	t.Helper()
	return begin(t, false)
}

func begin(t *testing.T, fromTemplate bool) Resources {
	t.Helper()
	if current == nil {
		t.Fatal("testenv did not start. Call testenv.Main in TestMain.")
	}

	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		t.Fatalf("The name of the test database is not available. Cause: %v", err)
	}
	name := "test_" + hex.EncodeToString(suffix)
	quoted := pgx.Identifier{name}.Sanitize()
	statement := "CREATE DATABASE " + quoted //nolint:gosec // DDL cannot use parameters. pgx.Identifier quotes the name.
	if fromTemplate {
		statement += " TEMPLATE " + pgx.Identifier{templateDB}.Sanitize()
	}
	if _, err := current.admin.ExecContext(t.Context(), statement); err != nil {
		t.Fatalf("The test database did not create. Cause: %v", err)
	}
	t.Cleanup(func() {
		// t.Context is already canceled here, so use a new context.
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if _, err := current.admin.ExecContext(ctx, "DROP DATABASE IF EXISTS "+quoted+" WITH (FORCE)"); err != nil {
			t.Errorf("The test database %s did not drop. Cause: %v", name, err)
		}
	})

	res := current.resources
	res.Postgres.DB = name

	client := OpenRedis(t, res.Redis)
	if err := client.FlushAll(t.Context()).Err(); err != nil {
		t.Fatalf("Redis did not empty. Cause: %v", err)
	}
	return res
}

// DatabaseExists tells if Postgres has a database with this name.
func DatabaseExists(t *testing.T, name string) bool {
	t.Helper()
	var exists bool
	err := current.admin.QueryRowContext(t.Context(),
		"SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = $1)", name).Scan(&exists)
	if err != nil {
		t.Fatalf("The database %s is not available. Cause: %v", name, err)
	}
	return exists
}

// TableExists tells if the database has a table with this name.
func TableExists(t *testing.T, db *sql.DB, table string) bool {
	t.Helper()
	var exists bool
	if err := db.QueryRowContext(t.Context(), "SELECT to_regclass($1) IS NOT NULL", table).Scan(&exists); err != nil {
		t.Fatalf("The table %s is not available. Cause: %v", table, err)
	}
	return exists
}

// OpenDB opens a connection pool that the test closes when it ends.
func OpenDB(t *testing.T, cfg config.Postgres) *sql.DB {
	t.Helper()
	db, err := sql.Open("pgx", cfg.DSN())
	if err != nil {
		t.Fatalf("The database did not open. Cause: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// OpenRedis opens a client that the test closes when it ends.
func OpenRedis(t *testing.T, cfg config.Redis) *redis.Client {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: cfg.Addr(), Password: cfg.Password, DB: cfg.CacheDB})
	t.Cleanup(func() { _ = client.Close() })
	return client
}
