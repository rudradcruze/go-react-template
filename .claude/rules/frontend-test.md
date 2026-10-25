---
paths:
  - "frontend/test/**"
  - "frontend/playwright.config.*"
---

# Frontend test rules

All frontend tests are in `frontend/test/`. Write each test before the code that it tests (see the constitution).

## Layout

```text
frontend/test/
├── unit/<feature>/   Vitest tests for components, hooks and functions
├── e2e/              Playwright tests
├── fixtures/         builders for test data
└── mocks/            MSW request handlers
```

## Unit tests

- Use Vitest and React Testing Library.
- Find elements as the user finds them: by role, label or text. Use a test ID only when no other query works.
- Use `@testing-library/user-event` for user actions.
- Send all API requests in unit tests to MSW. Type the MSW handlers with `src/api/schema.d.ts`.
- Set MSW to `onUnhandledRequest: "error"`, so that a request without a handler fails the test.

## Fixtures

- `fixtures/` has a builder for each API schema, for example `user({ email: "ann@example.com" })`.
- A builder gives a valid object with default values, typed from `src/api/schema.d.ts`. Each call gives unique values for unique fields.

## E2E tests

- Playwright runs against the real stack. E2E tests use no MSW.
- `make test-web-e2e` starts a separate Compose project with an empty database. Only the `web` service has a host port: 5655.
- Playwright uses `http://localhost:5655` as its `baseURL`.
- When the run ends, `make test-web-e2e` removes the project and its volumes with `docker compose down -v`, also when tests fail.
- Each test makes its own data with unique values. A test does not use data from a different test, so the tests can run in any order.
