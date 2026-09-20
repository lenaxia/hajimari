package handlers

import (
	"embed"
	"errors"
	"io"
	"mime"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	"github.com/toboshii/hajimari/frontend"
)

var ErrDir = errors.New("path is dir")

func tryRead(fs embed.FS, prefix, requestedPath string, w http.ResponseWriter) error {
	f, err := fs.Open(path.Join(prefix, requestedPath))
	if err != nil {
		return err
	}
	defer f.Close()

	stat, _ := f.Stat()
	if stat.IsDir() {
		return ErrDir
	}

	contentType := mime.TypeByExtension(filepath.Ext(requestedPath))
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", contentType)
	// SvelteKit emits content-hashed URLs under /_app/immutable/ — safe to
	// cache aggressively. Everything else (index.html, favicon, ...) must be
	// revalidated so browsers pick up new deployments instead of serving a
	// heuristically cached stale shell that references dead asset hashes.
	if strings.HasPrefix(requestedPath, "/_app/immutable/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	} else {
		w.Header().Set("Cache-Control", "no-cache")
	}
	_, err = io.Copy(w, f)
	return err
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {

	err := tryRead(frontend.BuildFs, "build", r.URL.Path, w)
	if err == nil {
		return
	}

	err = tryRead(frontend.BuildFs, "build", "index.html", w)
	if err != nil {
		panic(err)
	}
}
