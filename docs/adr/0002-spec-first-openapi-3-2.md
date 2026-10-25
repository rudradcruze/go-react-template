# Spec-first API with OpenAPI 3.2.0

The API contract is the hand-written file `backend/api/openapi.yaml`, in OpenAPI 3.2.0. We change the spec first, then the code. The frontend types come from the spec, and the E2E tests check the responses against it.

We use 3.2.0 because it is the newest version. Version 3.2.1 does not exist (October 2026). Not all of our tools fully support 3.2.0. Scalar shows its 3.2 support as "in progress", and `openapi-typescript` documents only 3.0 and 3.1. `pb33f/libopenapi-validator` supports 3.2. Before the spec uses a feature that is new in 3.2, make sure that Scalar and `openapi-typescript` accept it.

We did not generate the spec from Go code. A generated spec makes the backend the source of truth, and the frontend must then wait for the backend.
