package cart

import (
	"context"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/teatak/cart/v2/render"
)

type Engine struct {
	Router
	delims     render.Delims
	routers    map[string]*Router
	pool       sync.Pool
	paramsPool sync.Pool
	tree       *node

	NotFound HandlerFinal

	FuncMap  template.FuncMap
	Template *template.Template

	ForwardedByClientIP bool
	AppEngine           bool
	TrustedProxies      []string

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	OnRequest    func(*Context)
	OnResponse   func(*Context)
	ErrorHandler func(*Context, error)
}

var _ http.Handler = &Engine{}

func (e *Engine) allocateContext() *Context {
	return &Context{Response: &ResponseWriter{}}
}

func (e *Engine) getParams() *Params {
	ps, ok := e.paramsPool.Get().(*Params)
	if !ok {
		p := make(Params, 0, 20)
		return &p
	}
	*ps = (*ps)[:0]
	return ps
}

func (e *Engine) putParams(ps *Params) {
	if ps != nil {
		e.paramsPool.Put(ps)
	}
}

func (e *Engine) recycleContext(c *Context) {
	if ps := c.Params; ps != nil {
		e.putParams(ps)
	}
	e.pool.Put(c)
}

func (e *Engine) findRouter(absolutePath string) (*Router, bool) {
	router := e.routers[absolutePath]
	if router == nil {
		return nil, false
	}
	return router, true
}

func (e *Engine) getRouter(absolutePath string) (*Router, bool) {
	router := e.routers[absolutePath]
	find := true
	if router == nil {
		find = false
		router = &Router{
			Engine:  e,
			Path:    absolutePath,
			methods: make([]method, 0),
		}
	}
	return router, find
}

func (e *Engine) addRoute(router *Router) {
	if router.Path[0] != '/' {
		panic("Path must begin with '/' in path '" + router.Path + "'")
	}
	if e.tree == nil {
		e.tree = &node{}
	}
	//add router
	debugPrint("Add Router %s", router.Path)
	router.flatten() // Pre-calculate middleware chains
	e.routers[router.Path] = router
	e.tree.addRoute(router.Path, router)
	//if _, found := e.tree.findCaseInsensitivePath(router.Path, true); !found {
	//}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Context)
	defer e.recycleContext(c)

	c.reset(w, req)
	e.serveHTTP(c)
}

func (e *Engine) serve404(c *Context, path string) {
	// 404 error
	c.Router, _ = e.findRouter(path)
	if c.Response.Size() == -1 && c.Response.Status() == 200 {
		if e.NotFound != nil {
			if err := e.NotFound(c); err != nil && e.ErrorHandler != nil {
				e.ErrorHandler(c, err)
			}
		} else {
			c.ErrorHTML(404,
				"404 Not Found",
				"The page <b style='color:red'>"+path+"</b> is not found")
		}
	}
}

func (e *Engine) serveHTTP(c *Context) {
	if e.OnRequest != nil {
		e.OnRequest(c)
	}
	if e.OnResponse != nil {
		defer e.OnResponse(c)
	}
	path := c.Request.URL.Path
	httpMethod := c.Request.Method

	final404 := func() {
		e.serve404(c, path)
	}

	if root := e.tree; root != nil {
		if r, ps, tsr := root.getValue(path, e.getParams); r != nil {
			router := r.(*Router)
			c.Router = router
			c.Params = ps

			// Get pre-composed handler
			var handler HandlerCompose
			if h, ok := router.flattenHandlers[httpMethod]; ok {
				handler = h
			} else if h, ok := router.flattenHandlers["ANY"]; ok {
				handler = h
			}

			if handler != nil {
				handler(c, func() {})()
			} else {
				// try middleware only
				if router.composed != nil {
					router.composed(c, final404)()
				} else {
					final404()
				}
			}
			c.Response.WriteHeaderFinal()
			return
		} else if httpMethod != "CONNECT" && path != "/" {
			code := 301 // Permanent redirect, request with GET method
			if httpMethod != "GET" {
				code = 307
			}
			if tsr {
				if len(path) > 1 && path[len(path)-1] == '/' {
					c.Request.URL.Path = path[:len(path)-1]
				} else {
					c.Request.URL.Path = path + "/"
				}
				http.Redirect(c.Response, c.Request, c.Request.URL.String(), code)
				return
			}
		}
	}
	//find / middleware
	r, composed := e.mixComposed(path)
	if composed != nil {
		c.Router = r
		composed(c, final404)()
	} else {
		final404()
	}
	c.Response.WriteHeaderFinal()
}

func (e *Engine) mixComposed(absolutePath string) (*Router, HandlerCompose) {
	path := absolutePath
	for {
		if pr, find := e.findRouter(path); find {
			return pr, pr.composed
		}

		// Check with trailing slash if not already present
		if path != "/" && path[len(path)-1] != '/' {
			if pr, find := e.findRouter(path + "/"); find {
				return pr, pr.composed
			}
		}

		if path == "/" {
			break
		}

		// find last slash
		i := strings.LastIndexByte(path, '/')
		if i <= 0 {
			path = "/"
		} else {
			path = path[:i]
		}
	}
	return nil, nil
}

func (e *Engine) init() {
	e.Router = Router{
		Path: "/",
	}
	e.Router.Engine = e
	e.pool.New = func() interface{} {
		return e.allocateContext()
	}
	e.tree = &node{}
	e.routers = make(map[string]*Router)

	e.ReadTimeout = 90 * time.Second
	e.WriteTimeout = 90 * time.Second
	e.IdleTimeout = 90 * time.Second
}

func (e *Engine) Server(addr ...string) (server *http.Server) {
	address := resolveAddress(addr)
	debugPrint("PID:%d HTTP on %s\n", os.Getpid(), address)
	server = &http.Server{
		Addr:         address,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
		IdleTimeout:  e.IdleTimeout,
	}
	return
}

func (e *Engine) ServerKeepAlive(addr ...string) (server *http.Server) {
	address := resolveAddress(addr)
	debugPrint("PID:%d HTTP on %s\n", os.Getpid(), address)
	server = &http.Server{
		Addr:    address,
		Handler: e,
	}
	return
}

func (e *Engine) Run(addr string) (server *http.Server, err error) {
	defer func() { debugError(err) }()
	debugPrint("PID:%d Listening and serving HTTP on %s\n", os.Getpid(), addr)
	server = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
		IdleTimeout:  e.IdleTimeout,
	}
	err = server.ListenAndServe()
	return
}

func (e *Engine) RunTLS(addr string, certFile string, keyFile string) (server *http.Server, err error) {
	defer func() { debugError(err) }()
	debugPrint("PID:%d Listening and serving HTTPS on %s\n", os.Getpid(), addr)
	server = &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
		IdleTimeout:  e.IdleTimeout,
	}
	err = server.ListenAndServeTLS(certFile, keyFile)
	return
}

func (e *Engine) RunGraceful(addr string) error {
	server := &http.Server{
		Addr:         addr,
		Handler:      e,
		ReadTimeout:  e.ReadTimeout,
		WriteTimeout: e.WriteTimeout,
		IdleTimeout:  e.IdleTimeout,
	}

	go func() {
		debugPrint("PID:%d Listening and serving HTTP on %s\n", os.Getpid(), addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			debugError(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	debugPrint("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		debugError(err)
		return err
	}
	debugPrint("Server exiting")
	return nil
}

func (engine *Engine) LoadHTMLGlob(pattern string) {

	templ := template.Must(template.New("").Delims(engine.delims.Left, engine.delims.Right).Funcs(engine.FuncMap).ParseGlob(pattern))
	engine.SetHTMLTemplate(templ)

}

func (engine *Engine) SetHTMLTemplate(templ *template.Template) {
	engine.Template = templ.Funcs(engine.FuncMap)
}

func (engine *Engine) SetFuncMap(funcMap template.FuncMap) {
	engine.FuncMap = funcMap
}
