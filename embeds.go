package main

import "embed"

//go:embed static
var staticFS embed.FS

//go:embed templates
var templatesFS embed.FS

//go:embed resources
var resourcesFS embed.FS
