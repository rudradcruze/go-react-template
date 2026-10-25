# 06: Frontend skeleton

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A developer starts the frontend with `make dev` and opens it on port 5651. The frontend has typed access to the API through types that come from the spec. The API docs are also available through the frontend dev server.

**Blocked by:** 01 (Backend skeleton with API docs)

**Status:** ready-for-agent

- [ ] Do this first: `make gen` generates TypeScript types from the OpenAPI 3.2.0 spec with `openapi-typescript`. If the tool does not accept the 3.2.0 spec, stop and report to the user.
- [ ] The repo is a pnpm workspace, and the frontend is a workspace package.
- [ ] The frontend uses React and TypeScript in strict mode on the Vite+ toolchain, with Tailwind CSS v4, React Router and TanStack Query.
- [ ] The API client uses `openapi-fetch` with the generated types.
- [ ] The Vite dev server uses port 5651 and does not move to a different port. It sends `/api`, `/docs` and `/openapi.yaml` to the API on port 5653.
- [ ] `make dev` starts the Vite dev server together with the API.
- [ ] Vitest, React Testing Library and MSW are installed and configured. MSW fails each test that sends a request without a handler.
- [ ] One unit test proves that the home page shows.
- [ ] `make test`, `make lint` and `make fmt` also run on the frontend.
