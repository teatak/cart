package cart

import (
	"sync"
	"net/http"
	"os"
	"time"
)

type Engine struct {
	Router
	routers			map[string]*Router	//saved routers
	engine			*Engine
	pool        	sync.Pool
	tree			*node 	//match trees
}

var _ http.Handler = &Engine{}

var server *http.Server

func (e *Engine) allocateContext() *Context {
	return &Context{}
}

func (e *Engine) findRouter(absolutePath string) bool {
	router := e.routers[absolutePath]
	if router == nil {
		return false
	}
	return true
}

func (e *Engine) getRouter(absolutePath string) *Router {
	router := e.routers[absolutePath]
	if router == nil {
		router = &Router{
			engine:e,
			basePath:absolutePath,
		}
	}
	return router
}

func (e *Engine) addRoute(router *Router) {
	if router.basePath[0] != '/' {
		panic("Path must begin with '/' in path '" + router.basePath + "'")
	}
	if e.tree == nil {
		e.tree = &node{}
	}
	if _, found := e.tree.findCaseInsensitivePath(router.basePath, true); !found {
		e.routers[router.basePath] = router
		e.tree.addRoute(router.basePath, router)
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Context)
	c.reset(w, req)
	e.serveHTTP(c)
	e.pool.Put(c)
}

func (e *Engine) serveHTTP(c *Context) {
	path := c.Request.URL.Path
	httpMethod := c.Request.Method
	if root := e.tree; root != nil {
		if r, ps, tsr := root.getValue(path); r != nil {
			router := r.(*Router)
			c.Router = router
			c.Params = ps
			final := func() {
				// 404 error
				if c.Response.Size() == -1 && c.Response.Status() == 200 {
					c.Response.WriteString("empty page")
				}
			}
			composed := router.Composed
			if composed!=nil {
				composed(c,final)()
				//(c,final)()
				c.Response.WriteHeaderNow()
			}
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
	final404 := func() {
		// 404 error
		c.Response.WriteHeader(404)
		c.Response.WriteString("404 error")
	}
	if e.findRouter("/") {
		composed := e.getRouter("/").Composed
		composed(c,final404)()

	} else {
		final404()
	}
}

/*
init new Engine
 */
func (e *Engine) init() {
	e.Router = Router{
		basePath: "/",
	}
	e.Router.engine = e
	e.pool.New = func() interface{} {
		return e.allocateContext()
	}
	e.tree = &node{}
	e.routers = make(map[string]*Router)
}
/*
Run the server
 */
func (e *Engine) Run(addr ...string) (err error) {
	defer func() { debugError(err) }()
	address := resolveAddress(addr)
	debugPrint("PID:%d Listening and serving HTTP on %s\n", os.Getpid(), address)

	server = &http.Server{
		Addr: address,
		Handler: e,
		ReadTimeout: time.Second * 90,
		ReadHeaderTimeout: time.Second * 90,
		WriteTimeout: time.Second * 90,
		IdleTimeout: time.Second * 90,
	}
	err = server.ListenAndServe()
	return
}

