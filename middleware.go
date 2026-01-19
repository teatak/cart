package cart

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
)

// RequestID returns a middleware that adds a unique ID to each request
func RequestID() Handler {
	var (
		prefix string
		count  uint64
	)

	// Generate a unique prefix for this instance
	b := make([]byte, 8)
	rand.Read(b)
	prefix = base64.RawURLEncoding.EncodeToString(b)

	return func(c *Context, next Next) {
		id := c.Request.Header.Get("X-Request-ID")
		if id == "" {
			id = fmt.Sprintf("%s-%d", prefix, atomic.AddUint64(&count, 1))
		}
		c.Set("request_id", id)
		c.Header("X-Request-ID", id)
		next()
	}
}

// CORSConfig defines the config for CORS middleware
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig is the default config for CORS middleware
var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
	AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
}

// CORS returns a middleware that handles Cross-Origin Resource Sharing
func CORS(config ...CORSConfig) Handler {
	cfg := DefaultCORSConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	origins := make(map[string]bool)
	for _, o := range cfg.AllowOrigins {
		origins[o] = true
	}
	allowAll := origins["*"]

	methods := strings.Join(cfg.AllowMethods, ", ")
	headers := strings.Join(cfg.AllowHeaders, ", ")
	expose := strings.Join(cfg.ExposeHeaders, ", ")
	maxAge := fmt.Sprintf("%d", cfg.MaxAge)

	return func(c *Context, next Next) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			next()
			return
		}

		if allowAll {
			c.Header("Access-Control-Allow-Origin", "*")
		} else if origins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			// Origin not allowed, proceed without setting CORS headers
			// or strict block? usually just don't set header.
			next()
			return
		}

		if cfg.AllowCredentials && !allowAll {
			c.Header("Access-Control-Allow-Credentials", "true")
		}

		if expose != "" {
			c.Header("Access-Control-Expose-Headers", expose)
		}

		if c.Request.Method == "OPTIONS" {
			c.Header("Access-Control-Allow-Methods", methods)
			c.Header("Access-Control-Allow-Headers", headers)
			if cfg.MaxAge > 0 {
				c.Header("Access-Control-Max-Age", maxAge)
			}
			c.Status(http.StatusNoContent)
			c.Abort()
			return
		}

		next()
	}
}
