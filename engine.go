package cart

import (
	"sync"
	"net/http"
)

type Engine struct {
	Router
	engine			*Engine
	pool        	sync.Pool
	trees			map[string]*node 	//match trees
}

var _ IRouter = &Engine{}
var _ http.Handler = &Engine{}

var server *http.Server

func (e *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := e.pool.Get().(*Context)
	c.reset(w, req)
	e.serveHTTP(c)
	e.pool.Put(c)
}

func (e *Engine) serveHTTP(c *Context) {

}
/*
New Engine use default middleware
 */
func (e *Engine) New() {

}

/*
Use attachs a global middleware to the router
 */
func (r *Engine) Use(relativePath string, handle Handler) {

}
/*
Run the server
 */
func (e *Engine) Run() {

}

