package docs

import (
	"embed"
)

// OpenAPIFS contains the embedded OpenAPI specification.
//
//go:embed openapi.yaml
var OpenAPIFS embed.FS

// OpenAPIPath is the relative path for the OpenAPI document.
const OpenAPIPath = "openapi.yaml"
