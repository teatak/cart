package cart

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

func Logger() Handler {
	return LoggerWithWriter(DefaultWriter)
}

func LoggerWithWriter(out io.Writer) Handler {
	isTerm := true

	if _, ok := out.(*os.File); !ok || disableColor {
		isTerm = false
	}

	return func(c *Context, next Next) {
		start := time.Now()
		path := c.Request.URL.Path
		next()
		end := time.Now()
		latency := end.Sub(start)
		method := c.Request.Method
		clientIP := c.ClientIP()
		statusCode := c.Response.Status()
		var statusColor, methodColor string
		if isTerm {
			statusColor = colorForStatus(statusCode)
			methodColor = colorForMethod(method)
		}
		slog.Info(fmt.Sprintf("|%s %3d %s| %13v | %15s |%s %7s %s| %s",
			statusColor, statusCode, reset,
			latency,
			clientIP,
			methodColor, method, reset,
			path))
	}
}

func colorForStatus(code int) string {
	switch {
	case code >= 200 && code < 300:
		return greenBg
	case code >= 300 && code < 400:
		return whiteBg
	case code >= 400 && code < 500:
		return yellowBg
	default:
		return redBg
	}
}

var methodColors = map[string]string{
	"GET":     blue,
	"POST":    cyan,
	"PUT":     yellow,
	"DELETE":  red,
	"PATCH":   green,
	"HEAD":    magenta,
	"OPTIONS": white,
}

func colorForMethod(method string) string {
	if color, ok := methodColors[method]; ok {
		return color
	}
	return reset
}

var gzipPool = sync.Pool{
	New: func() interface{} {
		return gzip.NewWriter(io.Discard)
	},
}

// Gzip returns a middleware that compresses the response using gzip
func Gzip() Handler {
	return func(c *Context, next Next) {
		if !strings.Contains(c.Request.Header.Get("Accept-Encoding"), "gzip") {
			next()
			return
		}

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")
		// Delete Content-Length because the length changes after compression
		c.Header("Content-Length", "")

		// Wrap ResponseWriter to write gzipped data
		oldWriter := c.Response.ResponseWriter
		gz := gzipPool.Get().(*gzip.Writer)
		gz.Reset(oldWriter)
		defer func() {
			gz.Close()
			gzipPool.Put(gz)
		}()

		c.Response.ResponseWriter = &gzipWriter{ResponseWriter: oldWriter, gz: gz}
		defer func() {
			c.Response.ResponseWriter = oldWriter
		}()

		next()
	}
}

type gzipWriter struct {
	http.ResponseWriter
	gz *gzip.Writer
}

func (g *gzipWriter) Write(b []byte) (int, error) {
	// If Content-Type is missing, detect and set it before writing
	if g.Header().Get("Content-Type") == "" {
		g.Header().Set("Content-Type", http.DetectContentType(b))
	}
	return g.gz.Write(b)
}

func (g *gzipWriter) Flush() {
	if g.gz != nil {
		_ = g.gz.Flush()
	}
	if f, ok := g.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (g *gzipWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, ok := g.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, http.ErrNotSupported
	}
	return hijacker.Hijack()
}

func (g *gzipWriter) Push(target string, opts *http.PushOptions) error {
	pusher, ok := g.ResponseWriter.(http.Pusher)
	if !ok {
		return http.ErrNotSupported
	}
	return pusher.Push(target, opts)
}
