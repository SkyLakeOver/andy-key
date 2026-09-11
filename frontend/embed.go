package frontend

import (
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"embed"
)

//go:embed templates/*.html templates/components/*.html
var TemplatesFS embed.FS

//go:embed static/css/main.css static/js/alpine.min.js static/js/htmx.min.js static/js/app.js
var StaticFS embed.FS

// StaticFileServer возвращает http.Handler для раздачи статических файлов с правильными MIME-типами
func StaticFileServer(root http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Путь внутри embedded FS включает "static/", поэтому не убираем префикс
		// URL: /static/css/main.css -> файл в embed: static/css/main.css
		path := strings.TrimPrefix(r.URL.Path, "/")

		// Определяем MIME-тип по расширению файла
		ext := filepath.Ext(path)
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}

		// Открываем и отдаём файл
		f, err := root.Open(path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if stat.IsDir() {
			http.NotFound(w, r)
			return
		}

		http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
	})
}
