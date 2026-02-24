package main

import "embed"

// frontendFS contains the built Vue SPA (frontend/dist/).
// Built by: make frontend
//
//go:embed all:frontend/dist
var frontendFS embed.FS

// docsFS contains the offline Redoc API documentation (docs/api-ui/).
// redoc.standalone.js is downloaded by: make redoc
//
//go:embed all:docs/api-ui
var docsFS embed.FS
