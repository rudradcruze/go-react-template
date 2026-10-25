# 01: Backend skeleton with API docs

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A developer starts the API with one command. The API reads its settings from environment variables and writes JSON logs. When the API docs are enabled, a developer opens the docs in a browser on port 5653 and sees the API contract. This ticket also makes the backend test layout and the commands for tests, lint and format.

**Blocked by:** None (can start immediately)

**Status:** ready-for-agent

- [ ] The API starts on port 5653 with `make dev`.
- [ ] The API reads all settings from environment variables into one typed config. An incorrect value stops the API with a clear error.
- [ ] The API writes JSON logs with `slog`.
- [ ] The API serves the OpenAPI 3.2.0 spec at `/openapi.yaml` and the Scalar page at `/docs`.
- [ ] Scalar shows the 3.2.0 spec without errors. If Scalar cannot show it, stop and report to the user.
- [ ] With `API_DOCS_ENABLED=false`, both docs routes send 404.
- [ ] With docs auth enabled, a request without correct credentials gets 401, and a request with correct credentials gets 200.
- [ ] With docs auth enabled and an empty `API_DOCS_USERNAME` or `API_DOCS_PASSWORD`, the API does not start, and it logs a clear error.
- [ ] Unit tests cover the config checks. E2E tests cover the docs routes.
- [ ] `make test`, `make test-e2e`, `make lint` and `make fmt` work.
