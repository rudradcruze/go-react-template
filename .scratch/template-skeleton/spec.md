# Template skeleton (parent issue)

**Status:** ready-for-agent

## Goal

Build the first working version of the template. When all tickets are merged, the template has a Go API, a worker and a React frontend. Together, they show the example User context from end to end: Registration, the User page and the Welcome email.

The rules in `.claude/rules/` and the ADRs in `docs/adr/` apply to all tickets.

## Tickets

| # | Ticket | Blocked by |
| --- | --- | --- |
| 01 | [Backend skeleton with API docs](issues/01-backend-skeleton-with-api-docs.md) | None |
| 02 | [Dev infrastructure and E2E database isolation](issues/02-dev-infrastructure-and-e2e-database-isolation.md) | 01 |
| 03 | [Registration API](issues/03-registration-api.md) | 02 |
| 04 | [Get a User through the cache](issues/04-get-user-through-cache.md) | 03 |
| 05 | [Welcome email job](issues/05-welcome-email-job.md) | 03 |
| 06 | [Frontend skeleton](issues/06-frontend-skeleton.md) | 01 |
| 07 | [Registration page](issues/07-registration-page.md) | 03, 06 |
| 08 | [User page](issues/08-user-page.md) | 04, 07 |
| 09 | [Full stack in Docker Compose](issues/09-full-stack-in-docker-compose.md) | 05, 06 |
| 10 | [Frontend E2E with Playwright](issues/10-frontend-e2e-with-playwright.md) | 08, 09 |

## Waves

The tickets in one wave can be implemented at the same time. A ticket can start when all its blockers are merged. It does not have to wait for the full previous wave.

| Wave | Tickets (parallel) | Starts after |
| --- | --- | --- |
| 1 | 01 | None |
| 2 | 02, 06 | 01 |
| 3 | 03 | 02 |
| 4 | 04, 05, 07 | 03 (07 also needs 06) |
| 5 | 08, 09 | 08 needs 04 and 07. 09 needs 05 and 06. |
| 6 | 10 | 08, 09 |

The critical path is 01 → 02 → 03 → 04 → 08 → 10. A delay on this path delays the full project.

## Merge order

| Order | Ticket | Reason |
| --- | --- | --- |
| 1 | 01 | All tickets need it. |
| 2 | 02 | It is on the critical path, so it goes before 06. |
| 3 | 06 | Rebase on 02. Both tickets change the Makefile. |
| 4 | 03 | It unblocks wave 4. |
| 5 | 04 | It is on the critical path. It changes the spec, the User use cases and the API wiring. |
| 6 | 05 | Rebase on 04. Both tickets change the User use cases and the API wiring. |
| 7 | 07 | Rebase on 04, then run `make gen`, because 04 changed the spec. |
| 8 | 08 | It needs 04 and 07. |
| 9 | 09 | It goes after 08, so the web image contains the User page. |
| 10 | 10 | It needs the full stack and all pages. |

## Merge rules

- Merge one ticket at a time, in the merge order.
- Before the merge, rebase the branch on `main`.
- When `main` has a spec change, run `make gen` after the rebase.
- Before the merge, all unit tests and E2E tests must pass.

## Done when

- [ ] All 10 tickets are merged.
- [ ] `docker compose up` starts the full stack, and a person can register a User at `localhost:5651`.
- [ ] `make test`, `make test-e2e` and `make test-web-e2e` pass.
