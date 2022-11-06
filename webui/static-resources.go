//go:build webui

// Package webui contains the web user interface for embedding
package webui

import "embed"

//go:embed "dist/*"
var Dist embed.FS
