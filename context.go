package cart

import (
	"net/http"
	"net/url"
	"github.com/gimke/cart/render"
)

type Param struct {
	Key   string
	Value string
}

type Params []Param

func (ps Params) Get(name string) (string, bool) {
	for _, entry := range ps {
		if entry.Key == name {
			return entry.Value, true
		}
	}
	return "", false
}

type Context struct {
	response 	responseWriter
	Request   	*http.Request
	Response    ResponseWriter

	Router		*Router
	Params   	Params

	Keys     	map[string]interface{}
}

/*
reset Con
 */
func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.response.reset(w)
	c.Response = &c.response
	c.Request = req
	c.Params = c.Params[0:0]
	c.Keys = nil
}

//RESPONSE
func bodyAllowedForStatus(status int) bool {
	switch {
	case status >= 100 && status <= 199:
		return false
	case status == 204:
		return false
	case status == 304:
		return false
	}
	return true
}

// Header is a intelligent shortcut for c.Writer.Header().Set(key, value)
func (c *Context) Header(key, value string) {
	if len(value) == 0 {
		c.Response.Header().Del(key)
	} else {
		c.Response.Header().Set(key, value)
	}
}

// GetHeader returns value from request headers
func (c *Context) GetHeader(key string) string {
	return c.requestHeader(key)
}

func (c *Context) requestHeader(key string) string {
	if values, _ := c.Request.Header[key]; len(values) > 0 {
		return values[0]
	}
	return ""
}

func (c *Context) SetCookie(
	name string,
	value string,
	maxAge int,
	path string,
	domain string,
	secure bool,
	httpOnly bool,
) {
	if path == "" {
		path = "/"
	}
	http.SetCookie(c.Response, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

func (c *Context) Cookie(name string) (string, error) {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return "", err
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val, nil
}

func (c *Context) Status(code int) {
	c.response.WriteHeader(code)
}

func (c *Context) Render(code int, r render.Render) {
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Response)
		c.Response.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Response); err != nil {
		panic(err)
	}
}

// HTML renders the HTTP template specified by its file name.
// It also updates the HTTP code and sets the Content-Type as "text/html".
// See http://golang.org/doc/articles/wiki/
func (c *Context) HTML(code int, name string, obj interface{}) {
	//instance := render.HTMLProduction{Template }
	//c.Render(code, instance)
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, render.JSON{Data: obj})
}