package cart

import (
	"github.com/gimke/cart/render"
	"html/template"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Engine struct {
	Router
	delims  render.Delims
	routers map[string]*Router //saved routers
	engine  *Engine
	pool    sync.Pool
	tree    *node //match trees

	NotFound HandlerFinal

	FuncMap  template.FuncMap
	Template *template.Template

	ForwardedByClientIP bool
	AppEngine           bool
}

var _ http.Handler = &Engine{}

var server *http.Server

func (e *Engine) allocateContext() *Context {
	return &Context{}
}

func getParams() Params {
	ps := make(Params, 0, 20)
	return ps
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
	e.routers[router.Path] = router
	e.tree.addRoute(router.Path, router)
	//if _, found := e.tree.findCaseInsensitivePath(router.Path, true); !found {
	//}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Context)
	c.reset(w, req)
	e.serveHTTP(c)
	e.pool.Put(c)
}

func (e *Engine) mixMethods(httpMethod string, r *Router) HandlerCompose {
	//http method find any
	var methods HandlerCompose
	if m, find := r.getMethod("ANY"); find {
		methods = compose(m)
	}
	if m, find := r.getMethod(httpMethod); find {
		if methods != nil {
			methods = compose(methods, m)
		} else {
			methods = compose(m)
		}
	}
	return methods
}

func (e *Engine) serveHTTP(c *Context) {
	path := c.Request.URL.Path
	httpMethod := c.Request.Method

	final404 := func() {
		// 404 error
		// make temp router
		c.Router, _ = e.getRouter(path)
		if c.Response.Size() == -1 && c.Response.Status() == 200 {
			if e.NotFound != nil {
				e.NotFound(c)
			} else {
				c.ErrorHTML(404,
					"404 Not Found",
					"The page <b style='color:red'>"+path+"</b> is not found")
				//c.String(404,"404 Not Found")
			}
		}
	}

	if root := e.tree; root != nil {
		if r, ps, tsr := root.getValue(path, getParams); r != nil {
			router := r.(*Router)
			c.Router = router
			c.Params = ps

			//methods
			methods := e.mixMethods(httpMethod, router)
			//middleware

			composed := router.composed
			if composed != nil && methods != nil {
				composed = compose(composed, methods)
			} else if composed == nil && methods != nil {
				composed = methods
			}
			if composed != nil {
				composed(c, final404)()
			} else {
				final404()
			}
			c.Response.WriteHeaderNow()
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
	c.Response.WriteHeaderNow()
}

func (e *Engine) mixComposed(absolutePath string) (*Router, HandlerCompose) {
	sp := strings.Split(absolutePath, "/")
	for i, _ := range sp {
		//find it's self first ..... last is root path / router
		tempPath := strings.Join(sp[0:len(sp)-i], "/")
		if tempPath == "" {
			tempPath = "/"
		}
		if pr, find := e.findRouter(tempPath); find {
			return pr, pr.composed
		}
		//auto add slash then find
		if tempPath[len(tempPath)-1] != '/' && tempPath != absolutePath && tempPath != "/" {
			tempPath = tempPath + "/"
			if pr, find := e.findRouter(tempPath); find {
				return pr, pr.composed
			}
		}

	}
	return nil, nil
}

/*
init new Engine
*/
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
}

func (e *Engine) Server(addr ...string) (server *http.Server) {
	address := resolveAddress(addr)
	debugPrint("PID:%d HTTP on %s\n", os.Getpid(), address)
	server = &http.Server{
		Addr:        address,
		Handler:     e,
		ReadTimeout: time.Second * 90,
		//ReadHeaderTimeout: time.Second * 90,
		WriteTimeout: time.Second * 90,
		//IdleTimeout: time.Second * 90,
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

/*
Run the server
*/
func (e *Engine) Run(addr ...string) (server *http.Server, err error) {
	defer func() { debugError(err) }()
	address := resolveAddress(addr)
	debugPrint("PID:%d Listening and serving HTTP on %s\n", os.Getpid(), address)

	server = &http.Server{
		Addr:        address,
		Handler:     e,
		ReadTimeout: time.Second * 90,
		//ReadHeaderTimeout: time.Second * 90,
		WriteTimeout: time.Second * 90,
		//IdleTimeout: time.Second * 90,
	}
	err = server.ListenAndServe()
	return
}

/*
RunTLS
*/
func (e *Engine) RunTLS(addr string, certFile string, keyFile string) (server *http.Server, err error) {
	defer func() { debugError(err) }()
	debugPrint("PID:%d Listening and serving HTTPS on %s\n", os.Getpid(), addr)
	server = &http.Server{
		Addr:        addr,
		Handler:     e,
		ReadTimeout: time.Second * 90,
		//ReadHeaderTimeout: time.Second * 90,
		WriteTimeout: time.Second * 90,
		//IdleTimeout: time.Second * 90,
	}
	err = server.ListenAndServeTLS(certFile, keyFile)
	return
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
