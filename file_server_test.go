package cart

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStaticServe(t *testing.T) {
	// Setup temporary directory with a file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "hello.txt")
	err := os.WriteFile(filePath, []byte("Hello Static"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	app := New()
	// Use /assets prefix for static files served from tmpDir
	app.Use("/assets", Static(tmpDir, true))
	// For single file, File() middleware can be used
	app.Use("/hello", File(filePath))

	// Test Static Directory
	req := httptest.NewRequest("GET", "/assets/hello.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200 for static file, got %d", w.Code)
	}
	if w.Body.String() != "Hello Static" {
		t.Errorf("Expected content 'Hello Static', got %s", w.Body.String())
	}

	// Test Static File
	reqFile := httptest.NewRequest("GET", "/hello", nil)
	wFile := httptest.NewRecorder()
	app.ServeHTTP(wFile, reqFile)
	if wFile.Code != 200 {
		t.Errorf("Expected 200 for static file alias, got %d", wFile.Code)
	}
	if wFile.Body.String() != "Hello Static" {
		t.Errorf("Expected content 'Hello Static', got %s", wFile.Body.String())
	}
}

func TestStaticServeMissing(t *testing.T) {
	tmpDir := t.TempDir()
	app := New()
	app.Use("/assets", Static(tmpDir, true))

	req := httptest.NewRequest("GET", "/assets/missing.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 404 {
		t.Errorf("Expected 404 for missing file, got %d", w.Code)
	}
}

func TestFavicon(t *testing.T) {
	tmpDir := t.TempDir()
	iconPath := filepath.Join(tmpDir, "favicon.ico")
	err := os.WriteFile(iconPath, []byte("icon"), 0644)
	if err != nil {
		t.Fatal(err)
	}

	app := New()
	app.Use("/favicon.ico", Favicon(iconPath))

	// Test Request for favicon
	req := httptest.NewRequest("GET", "/favicon.ico", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200 for favicon, got %d", w.Code)
	}
	if w.Body.String() != "icon" {
		t.Errorf("Expected content 'icon', got %s", w.Body.String())
	}

	// Test Request for other path (should be ignored by favicon middleware if it was global,
	// but here it is mounted on specific path. However, Favicon middleware checks path internally too)
	// Let's mount it on root to test the internal check if we wanted, or just test standard usage.
	// The middleware implementation:
	// if c.Request.URL.Path != "/favicon.ico" { next(); return }
	// So it is safe to use globally or locally.

	app2 := New()
	app2.Use("/", Favicon(iconPath))
	// Register route for /other to test non-favicon access
	app2.Route("/other").GET(func(c *Context) error {
		c.String(200, "OK")
		return nil
	})

	req2 := httptest.NewRequest("GET", "/other", nil)
	w2 := httptest.NewRecorder()
	app2.ServeHTTP(w2, req2)
	if w2.Code != 200 || w2.Body.String() != "OK" {
		t.Errorf("Expected 200 OK for non-favicon request, got %d %s", w2.Code, w2.Body.String())
	}
}

func TestStaticFS(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "fs.txt")
	os.WriteFile(filePath, []byte("filesystem"), 0644)

	app := New()
	// StaticFS uses http.FileSystem
	app.Use("/fs", StaticFS("/fs", http.Dir(tmpDir)))

	req := httptest.NewRequest("GET", "/fs/fs.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200 for StaticFS, got %d", w.Code)
	}
	if w.Body.String() != "filesystem" {
		t.Errorf("Expected content 'filesystem', got %s", w.Body.String())
	}
}

func TestStaticDirectoryListing(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("test"), 0644)

	// Case 1: Listing Enabled
	// Use trailing slash to avoid TSR 301 redirect when listing directory
	app := New()
	app.Use("/list/", Static(tmpDir, true))

	req := httptest.NewRequest("GET", "/list/", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200 for directory listing, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "test.txt") {
		t.Errorf("Expected directory listing to contain 'test.txt'")
	}

	// Case 2: Listing Disabled
	app2 := New()
	app2.Use("/nolist/", Static(tmpDir, false))

	req2 := httptest.NewRequest("GET", "/nolist/", nil)
	w2 := httptest.NewRecorder()
	app2.ServeHTTP(w2, req2)

	// Note: Current implementation of neuteredReaddirFile returns nil, nil which results in 200 OK with empty listing.
	// Ideally it might return 403, but for now we verify that it doesn't list the files.
	if w2.Code != 200 {
		t.Errorf("Expected 200 for disabled directory listing (empty view), got %d", w2.Code)
	}
	if strings.Contains(w2.Body.String(), "test.txt") {
		t.Errorf("Expected directory listing to BE HIDDEN, but found 'test.txt'")
	}
}

func TestStaticFallback(t *testing.T) {
	tmpDir := t.TempDir()
	fallbackPath := filepath.Join(tmpDir, "index.html")
	os.WriteFile(fallbackPath, []byte("<h1>Fallback</h1>"), 0644)

	app := New()
	// Serve /app with fallback to index.html for 404s
	app.Use("/app", Static(tmpDir, false, fallbackPath))

	// Request existing file (we create one)
	os.WriteFile(filepath.Join(tmpDir, "main.js"), []byte("console.log('hi')"), 0644)
	req := httptest.NewRequest("GET", "/app/main.js", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 200 || w.Body.String() != "console.log('hi')" {
		t.Errorf("Expected normal file serve, got %d", w.Code)
	}

	// Request missing file -> Should return fallback
	req2 := httptest.NewRequest("GET", "/app/missing/route", nil)
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)

	if w2.Code != 200 {
		t.Errorf("Expected 200 (fallback) for missing file, got %d", w2.Code)
	}
	if w2.Body.String() != "<h1>Fallback</h1>" {
		t.Errorf("Expected fallback content, got %s", w2.Body.String())
	}
}
