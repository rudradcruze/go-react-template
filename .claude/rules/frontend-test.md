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
├── mocks/            MSW request handlers, and urls.ts with the URLs of the API
└── render.tsx        renderApp(path): renders the application with the router and a query client
```

## Unit tests

- Use Vitest and React Testing Library. Vitest comes with Vite+, so import `describe`, `it`, `expect` and `vi` from `vite-plus/test`. Run the tests with `vp test run`.
- Find elements as the user finds them: by role, label or text. Use a test ID only when no other query works.
- Use `@testing-library/user-event` for user actions.
- Send all API requests in unit tests to MSW. Type the MSW handlers with the types of `src/api/schema.d.ts`, for example `http.post<never, RegisterUserRequest, User>(url, ...)`.
- A test must wait until each request that its page sends has reached MSW. A request that arrives after the test ends gets no handler, and the setup reports it as a failure of the next test. Wait for the text that the response shows, not only for a heading or for an always-present region such as `role="alert"`.
- Test the failed responses that a proxy can send, not only the problem details of the API: an empty body, an HTML body, a JSON `null`, and a response with `Content-Length: 0`.
- `test/setup.ts` starts the mock server at the top of the file, before a test imports the app. The API client keeps the `fetch` function that exists when its module loads.
- A request without a handler gets a 500 response, and the `afterEach` hook fails the test, also when the app catches the error. A test that expects such a request takes the record from `unhandledRequests`.
- MSW 3 renamed `onUnhandledRequest` to `onUnhandledFrame`, and it ignores the old name without an error. With a function, MSW sends the request to the network unless the function throws.
- The test `test/unit/api/schema.test.ts` fails when `src/api/schema.d.ts` does not match the spec. Run `make gen` to correct it.

## Fixtures

- `fixtures/` has a builder for each API schema, for example `user({ email: "ann@example.com" })`.
- A builder gives a valid object with default values, typed from `src/api/schema.d.ts`. Each call gives unique values for unique fields.

## E2E tests

- Playwright runs against the real stack: nginx, the API, the worker, Postgres and Redis. E2E tests use no MSW.
- `make test-web-e2e` runs `scripts/test-web-e2e.sh`. The script starts the full stack (`compose.yaml` and `compose.e2e.yaml`) in the Compose project `go-react-template-e2e`, with an empty database and a new random password. The project reads neither `.env` nor the settings of your shell: the script uses `--env-file /dev/null` and clears them.
- Only the `web` service has a host port: 5655 (`E2E_WEB_PORT`). The development stack can run at the same time.
- Playwright uses `http://localhost:5655` as its `baseURL`. To run the tests against a stack that already runs, set `E2E_WEB_PORT` and run `pnpm --filter frontend run test:e2e`.
- If the script shows that a test failed or the stack did not start, it prints the end of the logs of the services before it removes the project. Do not run 2 copies of the script on one Docker host at the same time.
- When the run ends, the script removes the project and its volumes with `docker compose down --volumes`, also when a test fails (a `trap` on `EXIT`).
- Each test makes its own data with unique values (`test/e2e/support.ts`). A test does not use data from a different test, so the tests can run in any order, also at the same time. A test prepares data with `request.post` and tests the screen with `page`.
- `test/e2e/web-container.spec.ts` checks what nginx does: the proxy routes, the headers, the 413 answers and the health check.
