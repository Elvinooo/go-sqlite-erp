package main

import (
	"flag"
	"io"
	"log"
	"mime"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	listen := flag.String("listen", "0.0.0.0:5173", "listen address")
	webRoot := flag.String("web", "./web", "frontend dist directory")
	backend := flag.String("backend", "http://127.0.0.1:18080", "backend origin")
	flag.Parse()

	backendURL, err := url.Parse(*backend)
	if err != nil {
		log.Fatalf("invalid backend url: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "backend proxy error: "+err.Error(), http.StatusBadGateway)
	}

	root, err := filepath.Abs(*webRoot)
	if err != nil {
		log.Fatalf("invalid web root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, "index.html")); err != nil {
		log.Fatalf("index.html not found in %s: %v", root, err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			proxy.ServeHTTP(w, r)
			return
		}
		serveSPA(w, r, root)
	})

	log.Printf("frontend started on http://%s, web=%s, backend=%s", *listen, root, backendURL.String())
	if err := http.ListenAndServe(*listen, handler); err != nil {
		log.Fatal(err)
	}
}

func serveSPA(w http.ResponseWriter, r *http.Request, root string) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	path := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if path == "." || path == string(filepath.Separator) {
		path = "index.html"
	}
	fullPath := filepath.Join(root, path)
	absPath, err := filepath.Abs(fullPath)
	relPath, relErr := filepath.Rel(root, absPath)
	if err != nil || relErr != nil || relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if info, err := os.Stat(absPath); err != nil || info.IsDir() {
		absPath = filepath.Join(root, "index.html")
	}
	file, err := os.Open(absPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer file.Close()

	contentType := mime.TypeByExtension(filepath.Ext(absPath))
	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	if r.Method == http.MethodHead {
		return
	}
	_, _ = io.Copy(w, file)
}
