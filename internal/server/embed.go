package server

import "embed"

// FrontendFS and DocsFS are injected from main via embed.go at the root of the module.
// Using package-level vars allows the server package to be tested without embed.
var (
	FrontendFS embed.FS
	DocsFS     embed.FS
)
