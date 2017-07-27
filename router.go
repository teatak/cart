package cart


type (

	HandlerRoute func(IRouter)

	IRouter interface {
		Route(string, ...HandlerRoute) IRouter
		Handle(HandlerFinal) IRouter
		GET(HandlerFinal) IRouter
		POST(HandlerFinal) IRouter
		PUT(HandlerFinal) IRouter
		PATCH(HandlerFinal) IRouter
		HEAD(HandlerFinal) IRouter
		OPTIONS(HandlerFinal) IRouter
		DELETE(HandlerFinal) IRouter
		CONNECT(HandlerFinal) IRouter
		TRACE(HandlerFinal) IRouter
	}

	Router struct {
		engine *Engine
		parent IRouter
		basePath string
		root bool
	}
)

var _ IRouter = &Router{}

func (r *Router) Route(relativePath string, handles ...HandlerRoute) IRouter {
	return r
}

func (r *Router) Handle(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) GET(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) POST(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) PUT(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) PATCH(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) HEAD(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) OPTIONS(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) DELETE(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) CONNECT(handle HandlerFinal) IRouter {
	return r
}

func (r *Router) TRACE(handle HandlerFinal) IRouter {
	return r
}
