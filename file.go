package cart

import (
	"net/http"
	"strings"
	"os"
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

func Static(relativePath string, listDirectory bool) Handler {
	if strings.Contains(relativePath, ":") || strings.Contains(relativePath, "*") {
		panic("URL parameters can not be used when serving a static folder")
	}
	return func(c *Context, next Next) {
		fs := Dir(relativePath, listDirectory)
		prefix := c.Router.Path
		index := strings.LastIndex(prefix,"*")
		if index!=-1 {
			prefix = prefix[0:index]
		}
		fileServer := http.StripPrefix(prefix,http.FileServer(fs))
		fileServer.ServeHTTP(c.Response, c.Request)
		if(c.Response.Status() == 404) {
			c.Response.WriteHeader(200)	//reset status
			next()
		}

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