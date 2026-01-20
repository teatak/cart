package cart

import (
	"net/http"
	"net/url"
	"os"
	"strings"
)

func Favicon(relativePath string) Handler {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	return func(c *Context, next Next) {
		if c.Request.URL.Path != "/favicon.ico" {
			next()
			return
		} else {
			http.ServeFile(c.Response, c.Request, relativePath)
		}
	}
}

func File(relativePath string) Handler {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static file")
	}
	return func(c *Context, next Next) {
		http.ServeFile(c.Response, c.Request, relativePath)
	}
}

type fallbackResponseWriter struct {
	http.ResponseWriter
	code int
}

func (w *fallbackResponseWriter) WriteHeader(code int) {
	w.code = code
	if code != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(code)
	}
}

func (w *fallbackResponseWriter) Write(b []byte) (int, error) {
	if w.code == http.StatusNotFound {
		return len(b), nil
	}
	return w.ResponseWriter.Write(b)
}

func StripPrefixFallback(prefix string, fs http.FileSystem, listDirectory bool, fallback string) http.Handler {
	fileServer := http.FileServer(fs)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := strings.TrimPrefix(r.URL.Path, prefix)
		rp := strings.TrimPrefix(r.URL.RawPath, prefix)
		if len(p) < len(r.URL.Path) && (r.URL.RawPath == "" || len(rp) < len(r.URL.RawPath)) {
			r2 := new(http.Request)
			*r2 = *r
			r2.URL = new(url.URL)
			*r2.URL = *r.URL
			r2.URL.Path = p
			r2.URL.RawPath = rp

			fw := &fallbackResponseWriter{ResponseWriter: w}
			fileServer.ServeHTTP(fw, r2)

			if fw.code == http.StatusNotFound {
				if fallback != "" {
					if IsDebugging() {
						debugPrint("[WARNING] File not found: %s. Serving fallback: %s", p, fallback)
					}
					http.ServeFile(w, r, fallback)
				} else {
					http.NotFound(w, r)
				}
			}
		} else {
			http.NotFound(w, r)
		}
	})
}

// fallback can be empty if fallback is not empty http will render fallback path
func Static(relativePath string, listDirectory bool, fallback ...string) Handler {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	return func(c *Context, next Next) {
		fs := Dir(relativePath, listDirectory)
		prefix := c.Router.Path
		index := strings.LastIndex(prefix, "*")
		if index != -1 {
			prefix = prefix[0:index]
		}
		f := ""
		if len(fallback) > 0 {
			f = fallback[0]
		}
		fileServer := StripPrefixFallback(prefix, fs, listDirectory, f)
		fileServer.ServeHTTP(c.Response, c.Request)
	}
}

type (
	onlyfilesFS struct {
		fs http.FileSystem
	}
	neuteredReaddirFile struct {
		http.File
	}
)

// Dir returns a http.Filesystem that can be used by http.FileServer(). It is used internally
// in router.Static().
// if listDirectory == true, then it works the same as http.Dir() otherwise it returns
// a filesystem that prevents http.FileServer() to list the directory files.
func Dir(root string, listDirectory bool) http.FileSystem {
	fs := http.Dir(root)
	if listDirectory {
		return fs
	}
	return &onlyfilesFS{fs}
}

// Conforms to http.Filesystem
func (fs onlyfilesFS) Open(name string) (http.File, error) {
	f, err := fs.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return neuteredReaddirFile{f}, nil
}

// Overrides the http.File default implementation
func (f neuteredReaddirFile) Readdir(count int) ([]os.FileInfo, error) {
	// this disables directory listing
	return nil, nil
}

// StaticFS works like Static but with a custom http.FileSystem
func StaticFS(prefix string, fs http.FileSystem) Handler {
	actualPrefix := prefix
	if index := strings.LastIndex(prefix, "*"); index != -1 {
		actualPrefix = prefix[0:index]
	}
	return func(c *Context, next Next) {
		fileServer := http.StripPrefix(actualPrefix, http.FileServer(fs))
		fileServer.ServeHTTP(c.Response, c.Request)
	}
}
