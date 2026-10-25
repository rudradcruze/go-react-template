# syntax=docker/dockerfile:1
# This file builds the API, the worker and the migrate command into one image.
# A service in compose.yaml chooses the command to run with its entrypoint.

FROM golang:1.27.3-alpine AS build
WORKDIR /src
COPY . .
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/ ./cmd/api ./cmd/worker ./cmd/migrate

FROM alpine:3.22
# A user without special rights runs the process. The user ID is a number, so a platform can check it.
RUN adduser -D -u 10001 app
WORKDIR /app
COPY --from=build /out/ /app/
COPY migrations /app/migrations
USER 10001:10001
ENTRYPOINT ["/app/api"]
