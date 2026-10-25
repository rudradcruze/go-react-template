---
paths:
  - "backend/api-test/**"
  - "backend/.mockery.yaml"
---

# Backend test rules

All backend tests are in `backend/api-test/`. Write each test before the code that it tests (see the constitution).

## Layout

```text
backend/api-test/
├── unit/<context>/<layer>_test.go    unit tests
├── e2e/<context>/<feature>_test.go   E2E tests
├── fixtures/                         builders for test data
└── mocks/                            mocks from mockery (generated)
```

The tests are in different packages from the code. They use only exported identifiers, so they test behaviour through the public interface (see ADR 0003). When a test needs an unexported identifier, test the exported behaviour that uses it. Export an identifier only when production code needs it.

## Conventions

- Name each test `Test<Type>_<Method>_<Behaviour>`, for example `TestRegisterUser_Execute_RejectsTakenEmail`.
- Test one behaviour in each test. Use a table-driven test when the cases have the same steps.
- Use `require` for preconditions and `assert` for checks. Both come from Testify.
- Call `t.Parallel()` in unit tests. E2E tests run one after the other.
- Run unit tests with `make test`, and E2E tests with `make test-e2e`. For coverage, add `-coverpkg=./...`.

## Mocks

- A unit test replaces every port with a mock. A unit test does not touch Postgres, Redis or the network.
- mockery generates the mocks from the port interfaces into `mocks/`. Its configuration is `backend/.mockery.yaml`. Run `make gen` after a port changes. Do not edit the files in `mocks/`.
- Make each mock with its `NewMock<Port>(t)` constructor. The mock then checks its expectations when the test ends.
- Set expectations with `.EXPECT()`, and give exact arguments. Use `mock.Anything` only for a `context.Context`.

## Fixtures

- `fixtures/` has a builder for each entity, for example `fixtures.User(fixtures.WithEmail("ann@example.com"))`.
- A builder gives a valid object with default values. Each call gives unique values for unique fields, for example the email address.
- Unit tests use the builders to make domain objects. E2E tests use the save helpers, for example `fixtures.SaveUser(t, db)`, which save through the repository.

## E2E isolation

- Each E2E file starts with `//go:build e2e`.
- `TestMain` in each E2E package starts Postgres and Redis with testcontainers-go on random ports. It makes a template database and runs the goose migrations on it one time.
- Each test calls a helper that makes a new database from the template, with the name `test_<random>`. The helper drops the database in `t.Cleanup`, also when the test fails.
- The same helper removes all Redis data (`FLUSHALL`) before the test starts.
- Each test sends real HTTP requests to the real Gin engine, with no mocks.
- Each test checks each response against `backend/api/openapi.yaml` with `pb33f/libopenapi-validator`.
- When `TestMain` ends, it stops the containers. This also removes the template database.
