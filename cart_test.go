package cart

import (
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type mockResponseWriter struct{}

func (m *mockResponseWriter) Header() (h http.Header) {
	return http.Header{}
}

func (m *mockResponseWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (m *mockResponseWriter) WriteString(s string) (n int, err error) {
	return len(s), nil
}

func (m *mockResponseWriter) WriteHeader(int) {}

func Equal(t *testing.T, a interface{}, b interface{}, err string) {
	if a != b {
		t.Error(err)
	}
}

func handleAll(c *Context, next Next) {
	debugPrint("handleAll begin")
	next()
	debugPrint("handleAll end")
}

func handlePs(c *Context) error {
	debugPrint("handlePs begin")
	id, _ := c.Params.Get("id")
	debugPrint("id:%s", id)
	debugPrint("handlePs end")
	return nil
}

func TestEngine(t *testing.T) {
	c := New()
	c.Use("/", handleAll)
	c.Route("/:id").GET(handlePs)

	w := new(mockResponseWriter)
	req, _ := http.NewRequest("GET", "/", nil)
	reqa, _ := http.NewRequest("GET", "/123", nil)

	c.ServeHTTP(w, req)
	c.ServeHTTP(w, reqa)
}

// ==================== Default() Tests ====================

func TestDefault(t *testing.T) {
	app := Default()
	if app == nil {
		t.Fatal("Default() returned nil")
	}

	// Default should have Logger and RecoveryRender middlewares
	app.Route("/test").GET(func(c *Context) error {
		c.String(200, "default ok")
		return nil
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "default ok" {
		t.Errorf("Expected 'default ok', got %s", w.Body.String())
	}
}

func TestDefaultWithPanic(t *testing.T) {
	app := Default()
	app.Route("/panic").GET(func(c *Context) error {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	// RecoveryRender should catch the panic and return 500
	if w.Code != 500 {
		t.Errorf("Expected 500 after panic, got %d", w.Code)
	}
}

// ==================== HTML Template Tests ====================

func TestHTMLRender(t *testing.T) {
	app := New()

	// Create a template
	tpl := template.Must(template.New("test").Parse(`<h1>{{.Title}}</h1>`))
	app.SetHTMLTemplate(tpl)

	app.Route("/html").GET(func(c *Context) error {
		c.HTML(200, "test", H{"Title": "Hello"})
		return nil
	})

	req := httptest.NewRequest("GET", "/html", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "<h1>Hello</h1>") {
		t.Errorf("Expected HTML with Hello, got %s", w.Body.String())
	}
}

func TestLoadHTMLGlob(t *testing.T) {
	tmpDir := t.TempDir()
	tplPath := filepath.Join(tmpDir, "test.html")
	os.WriteFile(tplPath, []byte(`<p>{{.Name}}</p>`), 0644)

	app := New()
	app.LoadHTMLGlob(filepath.Join(tmpDir, "*.html"))

	app.Route("/glob").GET(func(c *Context) error {
		c.HTML(200, "test.html", H{"Name": "World"})
		return nil
	})

	req := httptest.NewRequest("GET", "/glob", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "<p>World</p>") {
		t.Errorf("Expected HTML with World, got %s", w.Body.String())
	}
}

func TestHTMLString(t *testing.T) {
	app := New()
	tpl := template.Must(template.New("str").Parse(`Name: {{.Name}}`))
	app.SetHTMLTemplate(tpl)

	app.Route("/str").GET(func(c *Context) error {
		html := c.HTMLString("str", H{"Name": "Test"})
		c.String(200, "Result: %s", html)
		return nil
	})

	req := httptest.NewRequest("GET", "/str", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "Name: Test") {
		t.Errorf("Expected 'Name: Test', got %s", w.Body.String())
	}
}

func TestLayoutHTML(t *testing.T) {
	app := New()
	tpl := template.Must(template.New("layout").Parse(`<body>{{.__CONTENT}}</body>`))
	template.Must(tpl.New("content").Parse(`<div>{{.Data}}</div>`))
	app.SetHTMLTemplate(tpl)

	app.Route("/layout").GET(func(c *Context) error {
		c.LayoutHTML(200, "layout", "content", H{"Data": "inner"})
		return nil
	})

	req := httptest.NewRequest("GET", "/layout", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "<div>inner</div>") {
		t.Errorf("Expected layout with inner content, got %s", w.Body.String())
	}
}

// ==================== Engine Server Tests ====================

func TestEngineServer(t *testing.T) {
	app := New()
	server := app.Server(":0")
	if server == nil {
		t.Fatal("Server() returned nil")
	}
	if server.Handler != app {
		t.Error("Server handler should be the engine")
	}
}

func TestEngineServerKeepAlive(t *testing.T) {
	app := New()
	server := app.ServerKeepAlive(":0")
	if server == nil {
		t.Fatal("ServerKeepAlive() returned nil")
	}
}

func TestEngineRun(t *testing.T) {
	app := New()
	app.Route("/").GET(func(c *Context) error {
		c.String(200, "ok")
		return nil
	})

	// Test Run in background and immediately stop
	go func() {
		time.Sleep(100 * time.Millisecond)
		// Server will be stopped by test timeout or next test
	}()

	// Just verify Run doesn't panic on invalid address
	_, err := app.Run("invalid-address")
	if err == nil {
		t.Error("Expected error for invalid address")
	}
}

func TestNotFound(t *testing.T) {
	app := New()
	app.NotFound = func(c *Context) error {
		c.String(404, "custom 404")
		return nil
	}

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 404 {
		t.Errorf("Expected 404, got %d", w.Code)
	}
	if w.Body.String() != "custom 404" {
		t.Errorf("Expected custom 404, got %s", w.Body.String())
	}
}

func TestErrorHandler(t *testing.T) {
	app := New()
	app.ErrorHandler = func(c *Context, err error) {
		c.String(500, "error: %v", err)
	}
	app.Route("/err").GET(func(c *Context) error {
		return &testError{msg: "something went wrong"}
	})

	req := httptest.NewRequest("GET", "/err", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "something went wrong") {
		t.Errorf("Expected error message, got %s", w.Body.String())
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

// ==================== File and Stream Tests ====================

func TestContextFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "download.txt")
	os.WriteFile(filePath, []byte("file content"), 0644)

	app := New()
	app.Route("/file").GET(func(c *Context) error {
		c.File(filePath)
		return nil
	})

	req := httptest.NewRequest("GET", "/file", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if w.Body.String() != "file content" {
		t.Errorf("Expected file content, got %s", w.Body.String())
	}
}

func TestContextStatic(t *testing.T) {
	tmpDir := t.TempDir()
	os.WriteFile(filepath.Join(tmpDir, "app.js"), []byte("console.log('hi')"), 0644)

	app := New()
	app.Route("/static/*filepath").GET(func(c *Context) error {
		c.Static(tmpDir, "/static/", false)
		return nil
	})

	req := httptest.NewRequest("GET", "/static/app.js", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestContextStream(t *testing.T) {
	app := New()
	app.Route("/stream").GET(func(c *Context) error {
		c.Header("Content-Type", "text/event-stream")
		count := 0
		c.Stream(func(w io.Writer) bool {
			count++
			if count > 3 {
				return false
			}
			return true
		})
		return nil
	})

	req := httptest.NewRequest("GET", "/stream", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

// ==================== ErrorHTML Tests ====================

func TestErrorHTML(t *testing.T) {
	app := New()
	app.Route("/error").GET(func(c *Context) error {
		c.ErrorHTML(500, "Server Error", "Something bad happened")
		return nil
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Server Error") {
		t.Errorf("Expected error title, got %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Something bad happened") {
		t.Errorf("Expected error content, got %s", w.Body.String())
	}
}

// ==================== AbortRender Tests ====================

func TestAbortRender(t *testing.T) {
	SetMode(DebugMode)
	defer SetMode(ReleaseMode)

	app := New()
	app.Route("/abortrender").GET(func(c *Context) error {
		c.AbortRender(500, "request info", "error message")
		return nil
	})

	req := httptest.NewRequest("GET", "/abortrender", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 500 {
		t.Errorf("Expected 500, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "error message") {
		t.Errorf("Expected error message, got %s", w.Body.String())
	}
}

// ==================== OnRequest/OnResponse Tests ====================

func TestOnRequestOnResponse(t *testing.T) {
	app := New()
	var requestCalled, responseCalled bool

	app.OnRequest = func(c *Context) {
		requestCalled = true
	}
	app.OnResponse = func(c *Context) {
		responseCalled = true
	}

	app.Route("/hooks").GET(func(c *Context) error {
		c.String(200, "ok")
		return nil
	})

	req := httptest.NewRequest("GET", "/hooks", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !requestCalled {
		t.Error("OnRequest was not called")
	}
	if !responseCalled {
		t.Error("OnResponse was not called")
	}
}

// ==================== SetFuncMap Tests ====================

func TestSetFuncMap(t *testing.T) {
	app := New()
	app.SetFuncMap(template.FuncMap{
		"upper": strings.ToUpper,
	})
	tpl := template.Must(template.New("func").Funcs(app.FuncMap).Parse(`{{upper .Name}}`))
	app.SetHTMLTemplate(tpl)

	app.Route("/func").GET(func(c *Context) error {
		c.HTML(200, "func", H{"Name": "hello"})
		return nil
	})

	req := httptest.NewRequest("GET", "/func", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "HELLO") {
		t.Errorf("Expected HELLO, got %s", w.Body.String())
	}
}
