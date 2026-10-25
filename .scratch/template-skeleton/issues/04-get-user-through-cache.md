# 04: Get a User through the cache

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A client gets a User by ID. The API reads the User from the Redis cache when the User is in the cache, and from Postgres when the User is not. If Redis fails, the API still returns the User from Postgres.

**Blocked by:** 03 (Registration API)

**Status:** ready-for-agent

- [ ] The spec defines `GET /api/v1/users/{id}` before the handler exists.
- [ ] A request for a registered User returns 200 with the User.
- [ ] A request for an ID that does not exist returns 404 as a problem detail.
- [ ] A request with an ID that is not a valid UUID returns 400 as a problem detail.
- [ ] The first read writes the User to the cache with a TTL of 5 minutes. The next read gets the User from the cache.
- [ ] When Redis fails, the API logs the error and returns the User from Postgres.
- [ ] Unit tests cover the cache hit, the cache miss and the Redis failure with mocks.
- [ ] An E2E test proves that the second read comes from the cache.
