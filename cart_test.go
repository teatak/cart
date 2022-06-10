package cart

import (
	"net/http"
	"testing"
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

//handlers
// func handle(c *Context) {
// 	debugPrint("handle")
// 	c.Response.Write([]byte(c.Request.URL.Path))
// }

func handleAll(c *Context, next Next) {
	debugPrint("handleAll begin")
	next()
	debugPrint("handleAll end")
}

func handlePs(c *Context) {
	debugPrint("handlePs begin")
	id, _ := c.Params.Get("id")
	debugPrint("id:%s", id)
	debugPrint("handlePs end")
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
