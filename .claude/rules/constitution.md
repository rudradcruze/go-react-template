# Constitution

These rules apply to all work in this repo. When a different rule or a task does not agree with this file, this file wins. Change this file only with the approval of the user, and record the change in an ADR.

1. **Test first.** Write a test that fails before you write the code that makes it pass. Then improve the code, and keep all tests green. Use the `tdd` skill.
2. **Write STE.** All human-readable text obeys `.claude/rules/writing.md`.
3. **The spec is the contract.** `backend/api/openapi.yaml` (OpenAPI 3.2.0) defines the API. Change the spec first, then the backend and the frontend.
4. **Keep the domain pure.** The domain layer contains only business rules. It imports no framework, ORM, cache, queue or logger.
5. **Mocks in unit tests, real infrastructure in E2E tests.** A unit test replaces every port with a mock. An E2E test uses real Postgres and real Redis, with no mocks.
6. **Configuration from the environment.** All settings come from environment variables. Each service uses its port from the port table in `architecture.md`, not the default port of the product.
7. **No secrets in the repo.** Put secrets only in `.env`, which git ignores. `.env.example` lists each variable with an empty or safe value.
