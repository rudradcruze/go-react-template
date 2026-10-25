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

- Use Vitest and React Testing Library. Vitest comes with Vite+, so import `describe`, `it`, `expect` and `vi` from `vite-plus/test`. Run the tests with `vp test run`.
- Find elements as the user finds them: by role, label or text. Use a test ID only when no other query works.
- Use `@testing-library/user-event` for user actions.
- Send all API requests in unit tests to MSW. Type the MSW handlers with `src/api/schema.d.ts`.
- `test/setup.ts` starts the mock server at the top of the file, before a test imports the app. The API client keeps the `fetch` function that exists when its module loads.
- A request without a handler gets a 500 response, and the `afterEach` hook fails the test, also when the app catches the error. A test that expects such a request takes the record from `unhandledRequests`.
- MSW 3 renamed `onUnhandledRequest` to `onUnhandledFrame`, and it ignores the old name without an error. With a function, MSW sends the request to the network unless the function throws.
- The test `test/unit/api/schema.test.ts` fails when `src/api/schema.d.ts` does not match the spec. Run `make gen` to correct it.

## Fixtures

- `fixtures/` has a builder for each API schema, for example `user({ email: "ann@example.com" })`.
- A builder gives a valid object with default values, typed from `src/api/schema.d.ts`. Each call gives unique values for unique fields.

## E2E tests

- Playwright runs against the real stack. E2E tests use no MSW.
- `make test-web-e2e` starts a separate Compose project with an empty database. Only the `web` service has a host port: 5655.
- Playwright uses `http://localhost:5655` as its `baseURL`.
- When the run ends, `make test-web-e2e` removes the project and its volumes with `docker compose down -v`, also when tests fail.
- Each test makes its own data with unique values. A test does not use data from a different test, so the tests can run in any order.
