package cart

import (
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	app := New()
	app.Use("/", CORS(CORSConfig{
		AllowOrigins:     []string{"http://example.com", "http://foo.com"},
		AllowCredentials: true,
	}))
	app.Route("/").GET(func(c *Context) error {
		c.String(200, "ok")
		return nil
	})

	// 1. Allowed Origin
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "http://example.com" {
		t.Errorf("Expected Allow-Origin: http://example.com, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" {
		t.Error("Expected Allow-Credentials: true")
	}

	// 2. Disallowed Origin
	req = httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Expected no Allow-Origin header for disallowed origin")
	}
}

func TestGzip(t *testing.T) {
	app := New()
	app.Use("/", Gzip())
	app.Route("/large").GET(func(c *Context) error {
		c.String(200, strings.Repeat("A", 1000))
		return nil
	})

	req := httptest.NewRequest("GET", "/large", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("Content-Encoding") != "gzip" {
		t.Error("Expected Content-Encoding: gzip")
	}
	if w.Header().Get("Vary") != "Accept-Encoding" {
		t.Error("Expected Vary: Accept-Encoding")
	}
	if w.Header().Get("Content-Length") != "" {
		t.Error("Expected Content-Length to be removed")
	}

	// Decode and verify body
	gr, err := gzip.NewReader(w.Body)
	if err != nil {
		t.Fatalf("Failed to create gzip reader: %v, Body: %q", err, w.Body.String())
	}
	defer gr.Close()
	body, err := io.ReadAll(gr)
	if err != nil {
		t.Fatalf("Failed to read gzip body: %v", err)
	}
	if string(body) != strings.Repeat("A", 1000) {
		t.Errorf("Gzip decoded body does not match. Got length: %d, Expected: 1000. Body start: %q", len(body), string(body[:min(len(body), 20)]))
	}
}

func TestClientIP(t *testing.T) {
	app := New()
	app.ForwardedByClientIP = true
	app.TrustedProxies = []string{"10.0.0.1"}

	app.Route("/ip").GET(func(c *Context) error {
		c.String(200, c.ClientIP())
		return nil
	})

	// 1. Trusted Proxy
	req := httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = "10.0.0.1:12345"
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Body.String() != "203.0.113.1" {
		t.Errorf("Expected ClientIP 203.0.113.1, got %s", w.Body.String())
	}

	// 2. Untrusted Proxy
	req = httptest.NewRequest("GET", "/ip", nil)
	req.RemoteAddr = "192.168.1.50:54321" // Not in TrustedProxies
	req.Header.Set("X-Forwarded-For", "203.0.113.1")
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// Should fallback to RemoteAddr (IP part)
	if w.Body.String() != "192.168.1.50" {
		t.Errorf("Expected ClientIP 192.168.1.50, got %s", w.Body.String())
	}
}
