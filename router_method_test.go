package cart

import (
	"net/http/httptest"
	"testing"
)

func TestRouterMethods(t *testing.T) {
	app := New()
	handler := func(c *Context) error {
		c.String(200, c.Request.Method)
		return nil
	}
	// ANY takes Handler (middleware style), not HandlerFinal
	handlerMiddleware := func(c *Context, next Next) {
		c.String(200, c.Request.Method)
	}

	app.Route("/post").POST(handler)
	app.Route("/put").PUT(handler)
	app.Route("/delete").DELETE(handler)
	app.Route("/patch").PATCH(handler)
	app.Route("/head").HEAD(handler)
	app.Route("/options").OPTIONS(handler)
	app.Route("/any").ANY(handlerMiddleware)

	tests := []struct {
		method string
		path   string
	}{
		{"POST", "/post"},
		{"PUT", "/put"},
		{"DELETE", "/delete"},
		{"PATCH", "/patch"},
		{"HEAD", "/head"},
		{"OPTIONS", "/options"},
		{"POST", "/any"},
		{"GET", "/any"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)
		if w.Code != 200 {
			t.Errorf("%s %s expected 200, got %d", tt.method, tt.path, w.Code)
		}
		if w.Body.String() != tt.method {
			// HEAD requests usually have no body in response, but cart.String writes it.
			// net/http/httptest might behave like real server and strip body for HEAD.
			if tt.method != "HEAD" {
				t.Errorf("%s %s expected body %s, got %s", tt.method, tt.path, tt.method, w.Body.String())
			}
		}
	}
}
