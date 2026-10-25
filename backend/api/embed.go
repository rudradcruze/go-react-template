// Package api holds the API contract.
package api

import _ "embed"

// Spec is the OpenAPI document of the API.
//
//go:embed openapi.yaml
var Spec []byte
