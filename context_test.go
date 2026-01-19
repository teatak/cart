package cart

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// Params
func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val, _ := ps.Get(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val, _ := ps.Get("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestContextReset(t *testing.T) {
	c := &Context{Response: &ResponseWriter{}}
	c.reset(nil, nil)
	if c.aborted {
		t.Error("Expected aborted to be false after reset")
	}
}

// ==================== Bind Tests ====================

func TestBindJSON(t *testing.T) {
	app := New()
	app.Route("/json").POST(func(c *Context) error {
		var data struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		if err := c.BindJSON(&data); err != nil {
			return err
		}
		c.JSON(200, H{"name": data.Name, "age": data.Age})
		return nil
	})

	body := `{"name":"John","age":30}`
	req := httptest.NewRequest("POST", "/json", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "John" {
		t.Errorf("Expected name John, got %v", resp["name"])
	}
}

func TestBindQuery(t *testing.T) {
	app := New()
	app.Route("/query").GET(func(c *Context) error {
		var data struct {
			Name string `form:"name"`
			Age  int    `form:"age"`
		}
		if err := c.BindQuery(&data); err != nil {
			return err
		}
		c.JSON(200, H{"name": data.Name, "age": data.Age})
		return nil
	})

	req := httptest.NewRequest("GET", "/query?name=Jane&age=25", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["name"] != "Jane" {
		t.Errorf("Expected name Jane, got %v", resp["name"])
	}
}

func TestBindForm(t *testing.T) {
	app := New()
	app.Route("/form").POST(func(c *Context) error {
		var data struct {
			Name string `form:"name"`
			Age  int    `form:"age"`
		}
		if err := c.BindForm(&data); err != nil {
			return err
		}
		c.JSON(200, H{"name": data.Name, "age": data.Age})
		return nil
	})

	form := url.Values{}
	form.Add("name", "Bob")
	form.Add("age", "40")
	req := httptest.NewRequest("POST", "/form", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
}

func TestBind(t *testing.T) {
	app := New()
	app.Route("/bind").POST(func(c *Context) error {
		var data struct {
			Name string `json:"name" form:"name"`
		}
		if err := c.Bind(&data); err != nil {
			c.String(400, "bind error: %v", err)
			return nil
		}
		c.String(200, "Hello %s", data.Name)
		return nil
	})

	// Test JSON binding
	body := `{"name":"Alice"}`
	req := httptest.NewRequest("POST", "/bind", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if !strings.Contains(w.Body.String(), "Alice") {
		t.Errorf("Expected Alice in response, got %s", w.Body.String())
	}

	// Test Form binding
	form := url.Values{}
	form.Add("name", "Charlie")
	req2 := httptest.NewRequest("POST", "/bind", strings.NewReader(form.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if !strings.Contains(w2.Body.String(), "Charlie") {
		t.Errorf("Expected Charlie in response, got %s", w2.Body.String())
	}
}

// ==================== Param Tests ====================

func TestParamMethods(t *testing.T) {
	app := New()
	app.Route("/user/:id").GET(func(c *Context) error {
		id, ok := c.Param("id")
		if !ok {
			return nil
		}
		c.String(200, "id=%s", id)
		return nil
	})

	req := httptest.NewRequest("GET", "/user/123", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "id=123" {
		t.Errorf("Expected id=123, got %s", w.Body.String())
	}
}

func TestParamInt(t *testing.T) {
	app := New()
	app.Route("/item/:num").GET(func(c *Context) error {
		num, err := c.ParamInt("num")
		if err != nil {
			c.String(400, "error: %v", err)
			return nil
		}
		c.String(200, "num=%d", num)
		return nil
	})

	req := httptest.NewRequest("GET", "/item/42", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "num=42" {
		t.Errorf("Expected num=42, got %s", w.Body.String())
	}
}

func TestParamInt64(t *testing.T) {
	app := New()
	app.Route("/big/:num").GET(func(c *Context) error {
		num, err := c.ParamInt64("num")
		if err != nil {
			c.String(400, "error: %v", err)
			return nil
		}
		c.String(200, "num=%d", num)
		return nil
	})

	req := httptest.NewRequest("GET", "/big/9999999999", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "num=9999999999" {
		t.Errorf("Expected num=9999999999, got %s", w.Body.String())
	}
}

// ==================== Abort Tests ====================

func TestAbort(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context, next Next) {
		c.Abort()
		next()
	})
	app.Route("/abort").GET(func(c *Context) error {
		c.String(200, "should not reach")
		return nil
	})

	req := httptest.NewRequest("GET", "/abort", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if strings.Contains(w.Body.String(), "should not reach") {
		t.Error("Handler should not have been called after Abort()")
	}
}

func TestAbortWithStatus(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context, next Next) {
		c.AbortWithStatus(401)
	})
	app.Route("/auth").GET(func(c *Context) error {
		c.String(200, "ok")
		return nil
	})

	req := httptest.NewRequest("GET", "/auth", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 401 {
		t.Errorf("Expected 401, got %d", w.Code)
	}
}

// ==================== KV Store Tests ====================

func TestContextKV(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context, next Next) {
		c.Set("user", "admin")
		c.Set("count", 100)
		c.Set("active", true)
		c.Set("rate", 3.14)
		c.Set("timestamp", time.Now())
		c.Set("duration", time.Second*5)
		c.Set("tags", []string{"a", "b"})
		c.Set("meta", map[string]interface{}{"key": "value"})
		c.Set("settings", map[string]string{"theme": "dark"})
		c.Set("options", map[string][]string{"colors": {"red", "blue"}})
		next()
	})
	app.Route("/kv").GET(func(c *Context) error {
		if c.GetString("user") != "admin" {
			return nil
		}
		if c.GetInt("count") != 100 {
			return nil
		}
		if !c.GetBool("active") {
			return nil
		}
		if c.GetFloat64("rate") != 3.14 {
			return nil
		}
		if c.GetTime("timestamp").IsZero() {
			return nil
		}
		if c.GetDuration("duration") != time.Second*5 {
			return nil
		}
		if len(c.GetStringSlice("tags")) != 2 {
			return nil
		}
		if c.GetStringMap("meta")["key"] != "value" {
			return nil
		}
		if c.GetStringMapString("settings")["theme"] != "dark" {
			return nil
		}
		if len(c.GetStringMapStringSlice("options")["colors"]) != 2 {
			return nil
		}
		c.String(200, "all kv ok")
		return nil
	})

	req := httptest.NewRequest("GET", "/kv", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "all kv ok" {
		t.Errorf("Expected 'all kv ok', got %s", w.Body.String())
	}
}

func TestMustGet(t *testing.T) {
	app := New()
	app.Use("/", func(c *Context, next Next) {
		c.Set("exists", "value")
		next()
	})
	app.Route("/mustget").GET(func(c *Context) error {
		val := c.MustGet("exists")
		c.String(200, "val=%v", val)
		return nil
	})

	req := httptest.NewRequest("GET", "/mustget", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "val=value" {
		t.Errorf("Expected val=value, got %s", w.Body.String())
	}
}

// ==================== Cookie Tests ====================

func TestCookie(t *testing.T) {
	app := New()
	app.Route("/setcookie").GET(func(c *Context) error {
		c.SetCookie("session", "abc123", 3600, "/", "", false, true)
		c.String(200, "cookie set")
		return nil
	})
	app.Route("/getcookie").GET(func(c *Context) error {
		val, err := c.Cookie("session")
		if err != nil {
			c.String(400, "no cookie")
			return nil
		}
		c.String(200, "session=%s", val)
		return nil
	})

	// Set cookie
	req := httptest.NewRequest("GET", "/setcookie", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	cookies := w.Result().Cookies()
	if len(cookies) == 0 || cookies[0].Name != "session" {
		t.Error("Expected session cookie to be set")
	}

	// Get cookie
	req2 := httptest.NewRequest("GET", "/getcookie", nil)
	req2.AddCookie(&http.Cookie{Name: "session", Value: "xyz789"})
	w2 := httptest.NewRecorder()
	app.ServeHTTP(w2, req2)
	if w2.Body.String() != "session=xyz789" {
		t.Errorf("Expected session=xyz789, got %s", w2.Body.String())
	}
}

// ==================== Header Tests ====================

func TestHeader(t *testing.T) {
	app := New()
	app.Route("/header").GET(func(c *Context) error {
		c.Header("X-Custom", "test-value")
		c.SetHeader("X-Another", "another-value")
		auth := c.GetHeader("Authorization")
		c.String(200, "auth=%s", auth)
		return nil
	})

	req := httptest.NewRequest("GET", "/header", nil)
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Header().Get("X-Custom") != "test-value" {
		t.Error("Expected X-Custom header")
	}
	if w.Header().Get("X-Another") != "another-value" {
		t.Error("Expected X-Another header")
	}
	if w.Body.String() != "auth=Bearer token123" {
		t.Errorf("Expected auth header, got %s", w.Body.String())
	}
}

// ==================== Render Tests ====================

func TestJSONRender(t *testing.T) {
	app := New()
	app.Route("/json").GET(func(c *Context) error {
		c.JSON(200, H{"status": "ok"})
		return nil
	})

	req := httptest.NewRequest("GET", "/json", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if !strings.Contains(w.Header().Get("Content-Type"), "application/json") {
		t.Error("Expected JSON content type")
	}
}

func TestIndentedJSONRender(t *testing.T) {
	app := New()
	app.Route("/indented").GET(func(c *Context) error {
		c.IndentedJSON(200, H{"pretty": true})
		return nil
	})

	req := httptest.NewRequest("GET", "/indented", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Body.String(), "\n") {
		t.Error("Expected indented JSON with newlines")
	}
}

func TestXMLRender(t *testing.T) {
	app := New()
	app.Route("/xml").GET(func(c *Context) error {
		c.XML(200, H{"data": "test"})
		return nil
	})

	req := httptest.NewRequest("GET", "/xml", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.Contains(w.Header().Get("Content-Type"), "application/xml") {
		t.Error("Expected XML content type")
	}
}

func TestJSONPRender(t *testing.T) {
	app := New()
	app.Route("/jsonp").GET(func(c *Context) error {
		c.JSONP(200, "callback", H{"value": 123})
		return nil
	})

	req := httptest.NewRequest("GET", "/jsonp", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if !strings.HasPrefix(w.Body.String(), "callback(") {
		t.Errorf("Expected JSONP callback, got %s", w.Body.String())
	}
}

func TestRedirect(t *testing.T) {
	app := New()
	app.Route("/old").GET(func(c *Context) error {
		c.Redirect(301, "/new")
		return nil
	})

	req := httptest.NewRequest("GET", "/old", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 301 {
		t.Errorf("Expected 301, got %d", w.Code)
	}
	if w.Header().Get("Location") != "/new" {
		t.Error("Expected Location header")
	}
}

func TestDataRender(t *testing.T) {
	app := New()
	app.Route("/data").GET(func(c *Context) error {
		c.Data(200, "application/octet-stream", []byte{0x01, 0x02, 0x03})
		return nil
	})

	req := httptest.NewRequest("GET", "/data", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), []byte{0x01, 0x02, 0x03}) {
		t.Error("Expected binary data")
	}
}

// ==================== Context Method Tests ====================

func TestContextMethod(t *testing.T) {
	app := New()
	app.Route("/ctx").GET(func(c *Context) error {
		ctx := c.Context()
		if ctx == nil {
			c.String(500, "context nil")
			return nil
		}
		c.String(200, "ok")
		return nil
	})

	req := httptest.NewRequest("GET", "/ctx", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "ok" {
		t.Errorf("Expected ok, got %s", w.Body.String())
	}
}

func TestContentType(t *testing.T) {
	app := New()
	app.Route("/ct").POST(func(c *Context) error {
		ct := c.ContentType()
		c.String(200, "ct=%s", ct)
		return nil
	})

	req := httptest.NewRequest("POST", "/ct", nil)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Body.String() != "ct=application/json" {
		t.Errorf("Expected ct=application/json, got %s", w.Body.String())
	}
}

func TestStatus(t *testing.T) {
	app := New()
	app.Route("/status").GET(func(c *Context) error {
		c.Status(204)
		return nil
	})

	req := httptest.NewRequest("GET", "/status", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != 204 {
		t.Errorf("Expected 204, got %d", w.Code)
	}
}
