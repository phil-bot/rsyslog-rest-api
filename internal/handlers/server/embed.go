package server

import "embed"

//go:embed all:../../frontend/dist
var frontendFS embed.FS

//go:embed all:../../docs/api-ui
var docsFS embed.FS

func init() {
	FrontendFS = frontendFS
	DocsFS = docsFS
}
