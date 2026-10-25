# 03: Registration API

**Parent:** `.scratch/template-skeleton/spec.md`

**What to build:** A client registers a User through the API with an email address and a display name. The API saves the User and returns it. The API rejects a second Registration with the same email address, also when the letter case is different. All errors are clear problem details.

**Blocked by:** 02 (Dev infrastructure and E2E database isolation)

**Status:** ready-for-agent

- [ ] The spec defines `POST /api/v1/users`, with its request, its response and its problem responses, before the handler exists.
- [ ] A successful Registration returns 201 with the User: ID, email address, display name and creation time.
- [ ] The API trims the email address and changes it to lowercase before it saves the User.
- [ ] A Registration with an email address that is already registered, in any letter case, returns 409.
- [ ] A Registration that breaks a domain rule returns 422. For example: an empty display name, or an email address that is not valid.
- [ ] A request body that the API cannot parse returns 400.
- [ ] All errors are RFC 9457 problem details. Their `title` and `detail` text obeys the STE rule.
- [ ] A migration makes the users table with a unique index on the email address.
- [ ] User IDs are UUIDv7 values that the domain makes.
- [ ] `make gen` generates the port mocks with mockery.
- [ ] Fixture builders make valid Users with unique email addresses.
- [ ] Unit tests cover the domain and the use case with mocks. E2E tests cover each status code and check each response against the spec.
