package cart

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// MockGQLHandler simulates 99designs/gqlgen handler
type MockGQLHandler struct{}

func (h *MockGQLHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	// Simulate query processing
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	if strings.Contains(string(body), "error") {
		w.WriteHeader(http.StatusOK) // GraphQL typically returns 200 even for errors
		w.Write([]byte(`{"errors":[{"message":"simulated error"}]}`))
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"data":{"hello":"world"}}`))
}

// MockPlaygroundHandler simulates 99designs/gqlgen/graphql/playground handler
type MockPlaygroundHandler struct {
	Title string
}

func (h *MockPlaygroundHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Playground writes HTML
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html><html><head><title>` + h.Title + `</title></head><body>Playground</body></html>`))
}

// MockSubscriptionHandler mimics a handler that uses context for cancellation
type MockSubscriptionHandler struct{}

func (h *MockSubscriptionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Modern way: use r.Context().Done()
	select {
	case <-r.Context().Done():
		return
	case <-time.After(10 * time.Millisecond):
		w.Write([]byte("data: connected\n\n"))
	}
}

func TestGQLIntegration(t *testing.T) {
	app := New()
	gqlHandler := &MockGQLHandler{}
	playgroundHandler := &MockPlaygroundHandler{Title: "GraphiQL"}

	// Integration pattern: Wrap standard http.Handler
	app.Route("/query").POST(func(c *Context) error {
		gqlHandler.ServeHTTP(c.Response, c.Request)
		return nil
	})

	app.Route("/playground").GET(func(c *Context) error {
		playgroundHandler.ServeHTTP(c.Response, c.Request)
		return nil
	})

	// Test 1: Query Success
	req := httptest.NewRequest("POST", "/query", strings.NewReader(`{"query":"{hello}"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Query expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello") {
		t.Errorf("Unexpected body: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Errorf("Expected application/json, got %s", w.Header().Get("Content-Type"))
	}

	// Test 2: Playground Success
	reqPlay := httptest.NewRequest("GET", "/playground", nil)
	wPlay := httptest.NewRecorder()
	app.ServeHTTP(wPlay, reqPlay)

	if wPlay.Code != 200 {
		t.Errorf("Playground expected 200, got %d", wPlay.Code)
	}
	if !strings.Contains(wPlay.Body.String(), "<title>GraphiQL</title>") {
		t.Errorf("Unexpected playground body: %s", wPlay.Body.String())
	}
	if !strings.Contains(wPlay.Header().Get("Content-Type"), "text/html") {
		t.Errorf("Expected text/html, got %s", wPlay.Header().Get("Content-Type"))
	}
}

func TestGQLSubscriptionCompatibility(t *testing.T) {
	// Test subscription handler using modern context-based cancellation
	app := New()
	subHandler := &MockSubscriptionHandler{}

	app.Route("/sub").GET(func(c *Context) error {
		subHandler.ServeHTTP(c.Response, c.Request)
		return nil
	})

	req := httptest.NewRequest("GET", "/sub", nil)
	w := httptest.NewRecorder()

	// Create a channel to wait for completion to avoid race in test cleanup
	done := make(chan bool)
	go func() {
		app.ServeHTTP(w, req)
		done <- true
	}()

	select {
	case <-done:
		if w.Body.String() != "data: connected\n\n" {
			t.Errorf("Expected stream data, got %s", w.Body.String())
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for subscription mock")
	}
}
