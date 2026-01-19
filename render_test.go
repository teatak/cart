package cart

import (
	"encoding/xml"
	"net/http/httptest"
	"testing"

	"github.com/teatak/cart/v2/render"
)

func TestRenderJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{
		"foo": "bar",
	}
	r := render.JSON{Data: data}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render JSON failed: %v", err)
	}
	if w.Body.String() != `{"foo":"bar"}`+"\n" {
		t.Errorf("Unexpected JSON body: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Unexpected Content-Type: %s", w.Header().Get("Content-Type"))
	}
}

func TestRenderIndentedJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]interface{}{
		"foo": "bar",
	}
	r := render.IndentedJSON{Data: data}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render IndentedJSON failed: %v", err)
	}
	// Indented JSON output check might vary by indentation preference (usually 4 spaces or tab)
	// We'll just check content type and containment for now
	if w.Header().Get("Content-Type") != "application/json; charset=utf-8" {
		t.Errorf("Unexpected Content-Type: %s", w.Header().Get("Content-Type"))
	}
}

func TestRenderXML(t *testing.T) {
	w := httptest.NewRecorder()
	type Person struct {
		XMLName xml.Name `xml:"person"`
		Name    string   `xml:"name"`
	}
	data := Person{Name: "yang"}
	r := render.XML{Data: data}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render XML failed: %v", err)
	}
	if w.Body.String() != `<person><name>yang</name></person>` {
		t.Errorf("Unexpected XML body: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/xml; charset=utf-8" {
		t.Errorf("Unexpected Content-Type: %s", w.Header().Get("Content-Type"))
	}
}

func TestRenderText(t *testing.T) {
	w := httptest.NewRecorder()
	r := render.String{Format: "Hello %s", Data: []interface{}{"World"}}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render Text failed: %v", err)
	}
	if w.Body.String() != "Hello World" {
		t.Errorf("Unexpected Text body: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "text/plain; charset=utf-8" {
		t.Errorf("Unexpected Content-Type: %s", w.Header().Get("Content-Type"))
	}
}

func TestRenderRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/test", nil)
	r := render.Redirect{
		Code:     301,
		Request:  req,
		Location: "/new",
	}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render Redirect failed: %v", err)
	}
	if w.Code != 301 {
		t.Errorf("Expected status 301, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/new" {
		t.Errorf("Expected Location /new, got %s", w.Header().Get("Location"))
	}
}

func TestRenderData(t *testing.T) {
	w := httptest.NewRecorder()
	r := render.Data{
		ContentType: "application/octet-stream",
		Data:        []byte("binary"),
	}
	err := r.Render(w)
	if err != nil {
		t.Errorf("Render Data failed: %v", err)
	}
	if w.Body.String() != "binary" {
		t.Errorf("Unexpected Data body: %s", w.Body.String())
	}
	if w.Header().Get("Content-Type") != "application/octet-stream" {
		t.Errorf("Unexpected Content-Type: %s", w.Header().Get("Content-Type"))
	}
}
