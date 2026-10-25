# 02: Dev infrastructure and E2E database isolation

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A developer starts Postgres and Redis with one command, on the ports of the port table, and runs the migrations. The API connects to both at startup. E2E tests get real Postgres and Redis in containers. Each E2E test gets its own database, and the database is deleted when the test ends.

**Blocked by:** 01 (Backend skeleton with API docs)

**Status:** ready-for-agent

- [ ] `make up` starts Postgres on host port 5654 and Redis on host port 5652. `make down` stops them, and `make logs` shows their logs.
- [ ] `make migrate-up` and `make migrate-down` run the goose migrations.
- [ ] The API connects to Postgres through GORM with the pgx driver, and to Redis. If a connection fails at startup, the API stops with a clear error.
- [ ] E2E tests start Postgres and Redis with testcontainers-go on random ports. They do not use the dev stack.
- [ ] Each E2E package makes a template database and migrates it one time.
- [ ] Each E2E test gets a new database from the template. The helper drops the database in `t.Cleanup`, so the drop also occurs after a failed test.
- [ ] Redis is empty at the start of each E2E test.
- [ ] A self-test proves that the test database exists during the test and does not exist after the test.
