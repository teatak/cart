package cart

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/teatak/cart/render"
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
	response responseWriter
	Request  *http.Request
	Response ResponseWriter

	Router *Router
	Params *Params
	Keys   map[string]interface{}
}

/*
reset Con
*/
func (c *Context) reset(w http.ResponseWriter, req *http.Request) {
	c.response.reset(w)
	c.Response = &c.response
	c.Request = req
	c.Params = nil
	c.Router = nil
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

func (c *Context) Param(key string) (string, bool) {
	return c.Params.Get(key)
}

// AbortWithStatus calls `Abort()` and writes the headers with the specified status code.
// For example, a failed attempt to authenticate a request could use: context.AbortWithStatus(401).
func (c *Context) AbortWithStatus(code int) {
	c.Status(code)
	c.Response.WriteHeaderNow()
}

func (c *Context) AbortRender(code int, request string, err interface{}) {
	stack := stack(3)
	if IsDebugging() {
		content := fmt.Sprintf("<pre>%s\n%s\n%s</pre>", request, err, stack)
		c.ErrorHTML(code,
			"Internal Server Error",
			content)
	} else {
		content := fmt.Sprintf("<pre>%s</pre>", err)
		c.ErrorHTML(code,
			"Internal Server Error",
			content)
	}
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
	if values := c.Request.Header[key]; len(values) > 0 {
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

// ClientIP implements the best effort algorithm to return the real client IP, it parses
// X-Real-IP and X-Forwarded-For in order to work properly with reverse-proxies such us: nginx or haproxy.
// Use X-Forwarded-For before X-Real-Ip as nginx uses X-Real-Ip with the proxy's IP.
func (c *Context) ClientIP() string {
	if c.Router.Engine.ForwardedByClientIP {
		clientIP := c.requestHeader("X-Forwarded-For")
		if index := strings.IndexByte(clientIP, ','); index >= 0 {
			clientIP = clientIP[0:index]
		}
		clientIP = strings.TrimSpace(clientIP)
		if len(clientIP) > 0 {
			return clientIP
		}
		clientIP = strings.TrimSpace(c.requestHeader("X-Real-Ip"))
		if len(clientIP) > 0 {
			return clientIP
		}
	}

	if c.Router.Engine.AppEngine {
		if addr := c.Request.Header.Get("X-Appengine-Remote-Addr"); addr != "" {
			return addr
		}
	}

	if ip, _, err := net.SplitHostPort(strings.TrimSpace(c.Request.RemoteAddr)); err == nil {
		return ip
	}

	return ""
}

// ContentType returns the Content-Type header of the request.
func (c *Context) ContentType() string {
	return filterFlags(c.requestHeader("Content-Type"))
}

func (c *Context) Render(code int, r render.Render) {
	c.response.writeNow = true
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
	instance := render.HTML{Template: c.Router.Engine.Template, Name: name, Data: obj}
	c.Render(code, instance)
}

// LayoutHTML render layout html
func (c *Context) LayoutHTML(code int, layout, name string, obj interface{}) {
	html := c.HTMLString(name, obj)
	tmp := obj.(H)
	tmp["__CONTENT"] = template.HTML(html)
	c.Render(code, render.HTML{Template: c.Router.Engine.Template, Name: layout, Data: tmp})
}

// HTMLString render Template to String
func (c *Context) HTMLString(name string, obj interface{}) string {
	tpl := c.Router.Engine.Template
	var buf bytes.Buffer
	err := tpl.ExecuteTemplate(&buf, name, obj)
	if err != nil {
		return ""
	}
	html := buf.String()
	return html
}

// IndentedJSON serializes the given struct as pretty JSON (indented + endlines) into the response body.
// It also sets the Content-Type as "application/json".
// WARNING: we recommend to use this only for development purposes since printing pretty JSON is
// more CPU and bandwidth consuming. Use Context.JSON() instead.
func (c *Context) IndentedJSON(code int, obj interface{}) {
	c.Render(code, render.IndentedJSON{Data: obj})
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj interface{}) {
	c.Render(code, render.JSON{Data: obj})
}

// XML serializes the given struct as XML into the response body.
// It also sets the Content-Type as "application/xml".
func (c *Context) XML(code int, obj interface{}) {
	c.Render(code, render.XML{Data: obj})
}

// String writes the given string into the response body.
func (c *Context) String(code int, format string, values ...interface{}) {
	c.Render(code, render.String{Format: format, Data: values})
}

// Redirect returns a HTTP redirect to the specific location.
func (c *Context) Redirect(code int, location string) {
	c.Render(-1, render.Redirect{
		Code:     code,
		Location: location,
		Request:  c.Request,
	})
}

// Data writes some data into the body stream and updates the HTTP code.
func (c *Context) Data(code int, contentType string, data []byte) {
	c.Render(code, render.Data{
		ContentType: contentType,
		Data:        data,
	})
}

// File writes the specified file into the body stream in a efficient way.
func (c *Context) File(filepath string) {
	http.ServeFile(c.Response, c.Request, filepath)
}

// Static file
func (c *Context) Static(relativePath, prefix string, listDirectory bool) {
	fs := Dir(relativePath, listDirectory)

	fileServer := http.StripPrefix(prefix, http.FileServer(fs))
	_, nolisting := fs.(*onlyfilesFS)
	if nolisting {
		c.Response.WriteHeader(404)
	}
	fileServer.ServeHTTP(c.Response, c.Request)
}

func (c *Context) Stream(step func(w io.Writer) bool) {
	w := c.Response
	clientGone := w.CloseNotify()
	for {
		select {
		case <-clientGone:
			return
		default:
			keepOpen := step(w)
			w.Flush()
			if !keepOpen {
				return
			}
		}
	}
}

// ErrorHTML error
func (c *Context) ErrorHTML(code int, title, content string) {
	tplString := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>{{.Title}}</title>
    <style>
    .title {
		display: block;
    	font-size: 2em;
    	font-weight: bold;
    	margin: 22px 0;
    }
    .content {
        margin: 10px 0;
        padding: 10px;
        background: linen;
        font-size: 14px;
    	line-height: 150%;
    }
    .content pre {
    	padding: 0;
    	margin: 0;
        white-space: pre-wrap;
		white-space: -moz-pre-wrap;
		white-space: -pre-wrap;
		white-space: -o-pre-wrap;
		word-wrap: break-word;
		word-break: break-all;
    }
    footer {
    	text-align: center;
		margin: 20px 0;
    	padding: 10px 0;
	}
    footer span {

    }
	footer a {
		display: inline-block;
		vertical-align: middle;
    }
	.center {
		margin-top: 16px;
		display: flex;
		justify-content: center;
    }
	pre {
		font-size: 10pt;
    	font-family: "Courier New", Monospace;
    	white-space: pre;
    }
    </style>
</head>
<body>
<div class="title">{{.Title}}</div>
<div class="content">{{.Content}}</div>
<footer>
	<span>powered by cart</span>
	<a target="_blank" href="https://github.com/teatak/cart"><svg width="22" height="22" class="octicon octicon-mark-github" viewBox="0 0 16 16" version="1.1" aria-hidden="true"><path fill-rule="evenodd" d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z"></path></svg></a>
</footer>
</body>
</html>
	`
	tpl, err := template.New("ErrorHTML").Parse(tplString)
	if err != nil {
		panic(err)
	}
	obj := H{
		"Title":   title,
		"Content": template.HTML(content),
	}
	c.Render(code, render.HTML{Template: tpl, Data: obj})
}

/************************************/
/******** METADATA MANAGEMENT********/
/************************************/

// Set is used to store a new key/value pair exclusively for this context.
// It also lazy initializes  c.Keys if it was not used previously.
func (c *Context) Set(key string, value interface{}) {
	if c.Keys == nil {
		c.Keys = make(map[string]interface{})
	}
	c.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (c *Context) Get(key string) (value interface{}, exists bool) {
	value, exists = c.Keys[key]
	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (c *Context) MustGet(key string) interface{} {
	if value, exists := c.Get(key); exists {
		return value
	}
	panic("Key \"" + key + "\" does not exist")
}

// GetString returns the value associated with the key as a string.
func (c *Context) GetString(key string) (s string) {
	if val, ok := c.Get(key); ok && val != nil {
		s, _ = val.(string)
	}
	return
}

// GetBool returns the value associated with the key as a boolean.
func (c *Context) GetBool(key string) (b bool) {
	if val, ok := c.Get(key); ok && val != nil {
		b, _ = val.(bool)
	}
	return
}

// GetInt returns the value associated with the key as an integer.
func (c *Context) GetInt(key string) (i int) {
	if val, ok := c.Get(key); ok && val != nil {
		i, _ = val.(int)
	}
	return
}

// GetInt64 returns the value associated with the key as an integer.
func (c *Context) GetInt64(key string) (i64 int64) {
	if val, ok := c.Get(key); ok && val != nil {
		i64, _ = val.(int64)
	}
	return
}

// GetFloat64 returns the value associated with the key as a float64.
func (c *Context) GetFloat64(key string) (f64 float64) {
	if val, ok := c.Get(key); ok && val != nil {
		f64, _ = val.(float64)
	}
	return
}

// GetTime returns the value associated with the key as time.
func (c *Context) GetTime(key string) (t time.Time) {
	if val, ok := c.Get(key); ok && val != nil {
		t, _ = val.(time.Time)
	}
	return
}

// GetDuration returns the value associated with the key as a duration.
func (c *Context) GetDuration(key string) (d time.Duration) {
	if val, ok := c.Get(key); ok && val != nil {
		d, _ = val.(time.Duration)
	}
	return
}

// GetStringSlice returns the value associated with the key as a slice of strings.
func (c *Context) GetStringSlice(key string) (ss []string) {
	if val, ok := c.Get(key); ok && val != nil {
		ss, _ = val.([]string)
	}
	return
}

// GetStringMap returns the value associated with the key as a map of interfaces.
func (c *Context) GetStringMap(key string) (sm map[string]interface{}) {
	if val, ok := c.Get(key); ok && val != nil {
		sm, _ = val.(map[string]interface{})
	}
	return
}

// GetStringMapString returns the value associated with the key as a map of strings.
func (c *Context) GetStringMapString(key string) (sms map[string]string) {
	if val, ok := c.Get(key); ok && val != nil {
		sms, _ = val.(map[string]string)
	}
	return
}

// GetStringMapStringSlice returns the value associated with the key as a map to a slice of strings.
func (c *Context) GetStringMapStringSlice(key string) (smss map[string][]string) {
	if val, ok := c.Get(key); ok && val != nil {
		smss, _ = val.(map[string][]string)
	}
	return
}
