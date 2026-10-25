# 09: Full stack in Docker Compose

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A person starts the full application with `docker compose up` and uses it on port 5651. The stack runs in the same way as in deployment.

**Blocked by:** 05 (Welcome email job), 06 (Frontend skeleton)

**Status:** ready-for-agent

- [ ] The API and the worker have multi-stage Docker images.
- [ ] The web image builds the frontend. In the web container, nginx serves the frontend and sends `/api`, `/docs` and `/openapi.yaml` to the API.
- [ ] A one-time migrate service runs the goose migrations before the API and the worker start.
- [ ] The Compose file holds the full stack: Postgres, Redis, migrate, API, worker and web. Postgres and Redis have no host ports.
- [ ] The web service uses port 5651, and the API uses port 5653.
- [ ] `make up` still starts only the development infrastructure.
- [ ] After `docker compose up`, a client registers a User through `localhost:5651/api/v1/users`, and the worker logs the Welcome email.
