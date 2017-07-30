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

func (r *Router) getMethod(httpMethod string) (HandlerCompose, bool) {
	for _, entry := range r.Methods {
		if entry.Key == httpMethod {
			return entry.HandlerCompose, true
		}
	}
	return nil, false
}

func (r *Router) mixComposed(absolutePath string) HandlerCompose {
	sp := strings.Split(absolutePath,"/")
	for i, _ := range sp {
		//find it's self first ..... last is root path / router
		tempPath := strings.Join(sp[0:len(sp)-i], "/")
		if tempPath == "" {
			tempPath = "/"
		}
		if pr, find := r.engine.findRouter(tempPath); find {
			return pr.Composed
		}
	}
	return nil
}

func (r *Router) use(absolutePath string, handler HandlerCompose) *Router {
	next := r.engine.getRouter(absolutePath)
	if composed := r.mixComposed(absolutePath); composed!=nil {
		next.Composed = compose(composed, handler)
	} else {
		next.Composed = compose(handler)
	}
	r.engine.addRoute(next)
	return next
}

func (r *Router) handle(httpMethod, absolutePath string, handler HandlerCompose) *Router {
	next := r.engine.getRouter(absolutePath)
	if composed := r.mixComposed(absolutePath); composed!=nil {
		next.Composed = compose(composed)
	}
	method := Method{Key:httpMethod, HandlerCompose:handler}
	next.Methods = append(next.Methods, method)
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
	next := r.use(absolutePath, makeCompose(handles...))
	return next
}


func (r *Router) ANY(handler Handler) *Router {
	return r.handle("ANY",r.basePath,makeCompose(handler))
}

func (r *Router) Handle(httpMethod string, handler HandlerFinal) *Router {
	tempHandler := func(c *Context,nex Next) {
		handler(c);
	}
	return r.handle(httpMethod,r.basePath,makeCompose(tempHandler))
}

func (r *Router) GET(handler HandlerFinal) *Router {
	return r.Handle("GET",handler)
}

func (r *Router) POST(handler HandlerFinal) *Router {
	return r.Handle("POST",handler)
}

func (r *Router) PUT(handler HandlerFinal) *Router {
	return r.Handle("PUT",handler)
}

func (r *Router) PATCH(handler HandlerFinal) *Router {
	return r.Handle("PATCH",handler)
}

func (r *Router) HEAD(handler HandlerFinal) *Router {
	return r.Handle("HEAD",handler)
}

func (r *Router) OPTIONS(handler HandlerFinal) *Router {
	return r.Handle("OPTIONS",handler)
}

func (r *Router) DELETE(handler HandlerFinal) *Router {
	return r.Handle("DELETE",handler)
}

func (r *Router) CONNECT(handler HandlerFinal) *Router {
	return r.Handle("CONNECT",handler)
}

func (r *Router) TRACE(handler HandlerFinal) *Router {
	return r.Handle("TRACE",handler)
}

