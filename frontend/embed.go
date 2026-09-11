package frontend

import "embed"

//go:embed templates/*.html templates/components/*.html
var TemplatesFS embed.FS

//go:embed static/*
var StaticFS embed.FS
