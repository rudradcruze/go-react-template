# 10: Frontend E2E with Playwright

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A developer runs the browser tests with one command. The command starts a separate, temporary copy of the full stack with an empty database. It runs Playwright against the copy, and then deletes the copy.

**Blocked by:** 08 (User page), 09 (Full stack in Docker Compose)

**Status:** ready-for-agent

- [ ] `make test-web-e2e` starts a separate Compose project with an empty database. Only the web service has a host port: 5655.
- [ ] The command can run while the dev stack runs.
- [ ] Playwright runs against `http://localhost:5655` and uses no MSW.
- [ ] When the run ends, the command removes the project and its volumes, also when tests fail.
- [ ] One test registers a User and sees the User page.
- [ ] One test registers an email address that is already registered and sees the error.
- [ ] Each test makes its own data with unique values, so the tests can run in any order.
