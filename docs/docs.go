// Package docs embeds all OpenAPI specifications and interactive Swagger UI.
package docs

import (
	"embed"
)

// FS embeds all documentation files including the interactive Swagger UI and YAML specifications.
//go:embed index.html
//go:embed swagger.yaml
//go:embed modules/*.yaml
var FS embed.FS
