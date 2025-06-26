package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"yourapp/internal/models"
	"yourapp/internal/shared"
)

//go:embed templates/**/*.html
//go:embed templates/*.html
var html embed.FS

//go:embed static/**/*.js
//go:embed static/**/*.css
var static embed.FS

func RenderPage(w http.ResponseWriter, r *http.Request, page string, data any) {
	funcs := template.FuncMap{
		"csrfToken": func() string {
			ss, found := r.Context().Value(shared.SessionKey).(*models.Session)
			if !found {
				return ""
			}
			return ss.CsrfCode
		},
		"json": func(v interface{}) (template.HTML, error) {
			b, err := json.Marshal(v)
			if err != nil {
				return "", err
			}
			// Safely embed JSON into HTML attribute.
			// template.HTMLEscapeString would escape quotes,
			// but template.HTML is fine if the JSON itself is valid.
			return template.HTML(b), nil
		},
	}
	t := template.Must(template.New("").Funcs(funcs).ParseFS(html, "templates/layout.html", "templates/pages/"+page))
	if err := t.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
}

func Static(mux *http.ServeMux) {
	f, _ := fs.Sub(static, "static")
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServerFS(f)))
}
