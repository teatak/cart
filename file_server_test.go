package cart

import (
	"net/http/httptest"
	"os"
	"path/filepath"
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
	// For single file, File() middleware can be used (but test used StaticFile which doesn't exist on Engine)
	// Assuming purpose is to test serving a specific file
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
