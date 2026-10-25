# All backend tests in one black-box folder

Go usually keeps each `_test.go` file next to its code. We keep all backend tests in `backend/api-test/`, with the folders `unit/`, `e2e/`, `fixtures/` and `mocks/`, so that all tests are in one location. The tests can use only exported identifiers, so they test behaviour through the public interface, as our TDD rules require. The folder is in the backend Go module, so it can import `internal/` packages. Coverage needs `-coverpkg=./...`, because the tests and the code are in different packages.
