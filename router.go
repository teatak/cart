package cart

import (
	"strings"
)

type (

	HandlerRoute func(*Router)

	Method struct {
		Key string
		HandlerCompose
	}
	Router struct {
		engine 		*Engine
		basePath 	string
		Composed	HandlerCompose
		Methods 	[]Method
	}
)
func (r *Router) handle(httpMethod, absolutePath string, handles HandlerCompose) *Router {
	next := r.engine.getRouter(absolutePath)
	sp := strings.Split(absolutePath[0:len(absolutePath)],"/")
	findParent := false
	for i, _ := range sp {
		tempPath :=  strings.Join(sp[0:len(sp)-i],"/")
		if tempPath == ""  {
			tempPath = "/"
		}
		if(r.engine.findRouter(tempPath)) {
			findParent = true
			parentComposed := r.engine.getRouter(tempPath).Composed
			next.Composed = compose(parentComposed,handles)
			break;
		}
	}

	if(!findParent) {
		next.Composed = compose(handles)
	}
	r.engine.addRoute(next)
	return next
}

func (r *Router) Route(relativePath string, handles ...HandlerRoute) *Router {
	absolutePath := joinPaths(r.basePath, relativePath)
	next := r.engine.getRouter(absolutePath)
	for _, handle := range handles {
		handle(next)
	}
	return next
}

func (r *Router) Use(relativePath string, handles ...Handler) *Router {
	absolutePath := joinPaths(r.basePath, relativePath)
	next := r.handle("ANY",absolutePath,makeCompose(handles...))
	return next
}

func (r *Router) GET(handle HandlerFinal) *Router {
	return r
}

func (r *Router) POST(handle HandlerFinal) *Router {
	return r
}

func (r *Router) PUT(handle HandlerFinal) *Router {
	return r
}

func (r *Router) PATCH(handle HandlerFinal) *Router {
	return r
}

func (r *Router) HEAD(handle HandlerFinal) *Router {
	return r
}

func (r *Router) OPTIONS(handle HandlerFinal) *Router {
	return r
}

func (r *Router) DELETE(handle HandlerFinal) *Router {
	return r
}

func (r *Router) CONNECT(handle HandlerFinal) *Router {
	return r
}

func (r *Router) TRACE(handle HandlerFinal) *Router {
	return r
}

