package frontend

import "embed"

//go:embed templates/*.html templates/components/*.html
var TemplatesFS embed.FS

//go:embed static/css/main.css static/js/alpine.min.js static/js/htmx.min.js static/js/app.js
var StaticFS embed.FS
