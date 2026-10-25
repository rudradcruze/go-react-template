---
paths:
  - "backend/api/**"
  - "backend/cmd/api/**"
  - "backend/internal/*/http/**"
  - "backend/internal/platform/server/**"
---

# API rules

## Spec first

- `backend/api/openapi.yaml` is the contract. Write the change in the spec before you write the handler.
- Use OpenAPI 3.2.0. Do not use these 3.2 features: the QUERY method, `additionalOperations` and `itemSchema`. `openapi-typescript` ignores them without an error (see ADR 0002). Before you use another feature that is new in 3.2, make sure that Scalar and `openapi-typescript` accept it. Put shared schemas in `components/schemas`. Give each operation an `operationId` in camelCase.
- After a spec change, run `make gen` to generate the frontend types again.
- An E2E test checks each response against the spec. See `api-test.md`.

## Routes and JSON

- Put all API routes under `/api/v1`.
- Use camelCase for JSON field names.
- Use these status codes:
  - 200 for a read
  - 201 for a create
  - 204 for a response with no body
  - 400 for a request that the API cannot parse
  - 401 when the credentials are missing or not correct
  - 404 when the resource does not exist
  - 409 for a conflict
  - 413 when the request body is larger than 1 MiB
  - 415 when the content type of the request is not `application/json`
  - 422 when a domain rule rejects the request
  - 500 for an unexpected error

## Handlers

A handler does 3 things:

1. Parse the request into the use case input. Read a JSON body with `httpjson.Decode`, which applies the rules for all endpoints: the content type, the maximum size, no unknown field and no data after the JSON value.
2. Call one use case.
3. Write the result or the error as the response.

The business rules stay in the domain. Use Gin binding only to parse JSON. Put no business rules in `binding` tags.

## Errors

- Send errors as RFC 9457 problem details, with the `application/problem+json` content type and the fields `type`, `title`, `status` and `detail`. Write them with `platform/problem`.
- The `title` is the HTTP status phrase, for example "Conflict". RFC 9457 asks for it when `type` is `about:blank`. It is the one text of the API that is not written in STE.
- Write `title` and `detail` in STE. State the problem, then tell the user what to do.
- Each bounded context has one function in `http/` that changes its domain errors into problem details. It writes them with `platform/problem`.
- For an unexpected error, send status 500 with a general `detail`. Write the cause to the log, not to the response.

## API docs

The API serves its docs with Scalar.

- `GET /docs` serves an HTML page that loads `@scalar/api-reference` from a CDN, with a pinned version.
- `GET /openapi.yaml` serves the spec. The binary contains the spec through `go:embed`.
- `API_DOCS_ENABLED=false` makes both routes send 404. This is the default value.
- `API_DOCS_AUTH_ENABLED=true` protects both routes with HTTP Basic auth. Compare the credentials with `crypto/subtle.ConstantTimeCompare`.
- When the docs and their auth are enabled, and `API_DOCS_USERNAME` or `API_DOCS_PASSWORD` is empty, the API does not start. It logs a clear error.
