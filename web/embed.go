package web

import (
	"embed"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"strings"
)

//go:embed all:docs/dist
var docsFS embed.FS

type Router interface {
	HandleFunc(pattern string, handler http.HandlerFunc)
	Mount(pattern string, handler http.Handler)
}

func DocsApp() WebApp { return NewWebApp("docs", docsFS, "docs/dist", "/docs/") }

type WebApp struct {
	name    string
	l       *slog.Logger
	fs      fs.FS
	urlBase string
}

func NewWebApp(name string, app fs.FS, subDir string, urlBase string) WebApp {
	subFS, err := fs.Sub(app, subDir)
	if err != nil {
		log.Fatal(err)
	}

	// Ensure urlBase ends with /
	urlBase = strings.TrimSuffix(urlBase, "/") + "/"

	return WebApp{
		name:    name,
		fs:      subFS,
		urlBase: urlBase,
		l:       slog.Default().With(slog.String("component", name)),
	}
}

func (wa WebApp) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	if path == "" {
		path = "index.html"
	}

	// Try alternative paths, including exact file.
	altSuffixes := []string{"", ".html", "/index.html"}
	for _, suffix := range altSuffixes {
		altPath := strings.TrimSuffix(path, "/") + suffix
		if f, err := fs.Stat(wa.fs, altPath); err == nil {
			if f.IsDir() {
				continue
			}

			http.ServeFileFS(w, r, wa.fs, altPath)

			return
		}
	}

	wa.l.Warn("File not found", slog.String("path", path))

	// File not found, redirect to base URL
	http.Redirect(w, r, wa.urlBase, http.StatusTemporaryRedirect)
}

func (wa WebApp) URLBase() string {
	return wa.urlBase
}

// Handler returns an http.Handler that serves the WebApp at the given path.
func (wa WebApp) Handler(path string) http.Handler {
	return http.StripPrefix(path, wa)
}

// Register registers the WebApp with the given ServeMux.
func (wa WebApp) Register(mux Router, l *slog.Logger) {
	wa.l = l.With(slog.String("app", wa.name), slog.String("urlBase", wa.urlBase), slog.String("component", "file-server"))
	wa.l.Info("Registering web app")

	// Redirect base without trailing slash to base with slash
	baseWithoutSlash := strings.TrimSuffix(wa.urlBase, "/")
	mux.HandleFunc(baseWithoutSlash, func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, wa.urlBase, http.StatusMovedPermanently)
	})

	// Mount the web app with prefix stripping
	mux.Mount(baseWithoutSlash, wa.Handler(wa.urlBase))
}
