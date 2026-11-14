// Package api expone el contrato OpenAPI y la UI de documentación embebidos.
package api

import _ "embed"

//go:embed openapi.yaml
var Spec []byte

//go:embed docs.html
var DocsHTML []byte

//go:embed favicon.svg
var FaviconSVG []byte
