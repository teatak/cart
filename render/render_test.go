package render

import (
	"html/template"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONRender(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"name": "test"}
	r := JSON{Data: data}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Error("Expected JSON content type")
	}
	if !strings.Contains(w.Body.String(), `"name":"test"`) {
		t.Errorf("Expected JSON body, got %s", w.Body.String())
	}
}

func TestJSONWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := JSON{}
	r.WriteContentType(w)

	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Error("Expected JSON content type")
	}
}

func TestIndentedJSONRender(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	r := IndentedJSON{Data: data}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Body.String(), "\n") {
		t.Error("Expected indented JSON with newlines")
	}
	if !strings.Contains(w.Body.String(), "    ") {
		t.Error("Expected indented JSON with spaces")
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	err := WriteJSON(w, map[string]int{"count": 42})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !strings.Contains(w.Body.String(), "42") {
		t.Error("Expected 42 in JSON")
	}
}

func TestXMLRender(t *testing.T) {
	w := httptest.NewRecorder()
	type Item struct {
		Name string `xml:"name"`
	}
	r := XML{Data: Item{Name: "test"}}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "application/xml") {
		t.Error("Expected XML content type")
	}
	if !strings.Contains(w.Body.String(), "<name>test</name>") {
		t.Errorf("Expected XML body, got %s", w.Body.String())
	}
}

func TestXMLWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := XML{}
	r.WriteContentType(w)

	if !strings.Contains(w.Header().Get("Content-Type"), "application/xml") {
		t.Error("Expected XML content type")
	}
}

func TestHTMLRender(t *testing.T) {
	w := httptest.NewRecorder()
	tpl := template.Must(template.New("test").Parse(`<h1>{{.Title}}</h1>`))
	r := HTML{Template: tpl, Name: "test", Data: map[string]string{"Title": "Hello"}}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Expected HTML content type")
	}
	if !strings.Contains(w.Body.String(), "<h1>Hello</h1>") {
		t.Errorf("Expected HTML body, got %s", w.Body.String())
	}
}

func TestHTMLRenderWithoutName(t *testing.T) {
	w := httptest.NewRecorder()
	tpl := template.Must(template.New("").Parse(`<p>{{.}}</p>`))
	r := HTML{Template: tpl, Data: "World"}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Body.String(), "<p>World</p>") {
		t.Errorf("Expected HTML body, got %s", w.Body.String())
	}
}

func TestHTMLWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := HTML{}
	r.WriteContentType(w)

	if !strings.Contains(w.Header().Get("Content-Type"), "text/html") {
		t.Error("Expected HTML content type")
	}
}

func TestStringRender(t *testing.T) {
	w := httptest.NewRecorder()
	r := String{Format: "Hello %s!", Data: []interface{}{"World"}}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !strings.Contains(w.Header().Get("Content-Type"), "text/plain") {
		t.Error("Expected plain text content type")
	}
	if w.Body.String() != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got %s", w.Body.String())
	}
}

func TestStringRenderWithoutData(t *testing.T) {
	w := httptest.NewRecorder()
	r := String{Format: "Static text"}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Body.String() != "Static text" {
		t.Errorf("Expected 'Static text', got %s", w.Body.String())
	}
}

func TestWriteString(t *testing.T) {
	w := httptest.NewRecorder()
	WriteString(w, "Format %d", []interface{}{123})
	if w.Body.String() != "Format 123" {
		t.Errorf("Expected 'Format 123', got %s", w.Body.String())
	}
}

func TestDataRender(t *testing.T) {
	w := httptest.NewRecorder()
	r := Data{ContentType: "application/octet-stream", Data: []byte{0x01, 0x02, 0x03}}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Error("Expected custom content type")
	}
	if len(w.Body.Bytes()) != 3 {
		t.Errorf("Expected 3 bytes, got %d", len(w.Body.Bytes()))
	}
}

func TestDataWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := Data{ContentType: "image/png"}
	r.WriteContentType(w)

	if w.Header().Get("Content-Type") != "image/png" {
		t.Error("Expected image/png content type")
	}
}

func TestRedirectRender(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/old", nil)
	r := Redirect{Code: 301, Request: req, Location: "/new"}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Code != 301 {
		t.Errorf("Expected 301, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/new" {
		t.Error("Expected Location header")
	}
}

func TestRedirectRender201(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/create", nil)
	r := Redirect{Code: 201, Request: req, Location: "/resource/1"}

	err := r.Render(w)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if w.Code != 201 {
		t.Errorf("Expected 201, got %d", w.Code)
	}
}

func TestRedirectInvalidCode(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for invalid redirect code")
		}
	}()

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	r := Redirect{Code: 200, Request: req, Location: "/new"}
	r.Render(w)
}

func TestRedirectWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := Redirect{}
	r.WriteContentType(w)
	// Should not panic, just a no-op
}

func TestWriteContentType(t *testing.T) {
	w := httptest.NewRecorder()
	writeContentType(w, []string{"custom/type"})
	if w.Header().Get("Content-Type") != "custom/type" {
		t.Error("Expected custom content type")
	}
}

// Test Delims struct
func TestDelims(t *testing.T) {
	d := Delims{Left: "{{", Right: "}}"}
	if d.Left != "{{" || d.Right != "}}" {
		t.Error("Expected default delims")
	}
}

// Test interface compliance
func TestRenderInterface(t *testing.T) {
	var _ Render = JSON{}
	var _ Render = IndentedJSON{}
	var _ Render = XML{}
	var _ Render = String{}
	var _ Render = Redirect{}
	var _ Render = Data{}
	var _ Render = HTML{}
}
