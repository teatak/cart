package cart

type (
	HandlerRoute func(*Router)

	method struct {
		key     string
		handler HandlerCompose
	}
	Router struct {
		Engine          *Engine
		Path            string
		composed        HandlerCompose
		methods         []method
		flattenHandlers map[string]HandlerCompose // Pre-calculated handlers per method
	}
)

func (r *Router) getMethod(httpMethod string) (HandlerCompose, bool) {
	for _, entry := range r.methods {
		if entry.key == httpMethod {
			return entry.handler, true
		}
	}
	return nil, false
}

func (r *Router) flatten() {
	if r.flattenHandlers == nil {
		r.flattenHandlers = make(map[string]HandlerCompose)
	}

	baseComposed := r.composed

	// 先获取 ANY handler
	var anyHandler HandlerCompose
	if mh, ok := r.getMethod("ANY"); ok {
		anyHandler = mh
	}

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD", "ANY"}
	for _, m := range methods {
		var handler HandlerCompose
		if mh, ok := r.getMethod(m); ok {
			handler = mh
		} else if m != "ANY" && anyHandler != nil {
			// 如果没有特定方法 handler 且有 ANY handler，使用 ANY handler
			handler = anyHandler
		}

		if baseComposed != nil {
			if handler != nil {
				r.flattenHandlers[m] = compose(baseComposed, handler)
			} else {
				// If no specific method handler, use base (middleware only)
				r.flattenHandlers[m] = baseComposed
			}
		} else if handler != nil {
			r.flattenHandlers[m] = handler
		}
	}
}

func (r *Router) use(absolutePath string, handler HandlerCompose) *Router {
	next, find := r.Engine.getRouter(absolutePath)
	if _, composed := r.Engine.mixComposed(absolutePath); composed != nil {
		next.composed = compose(composed, handler)
	} else {
		next.composed = compose(handler)
	}
	if !find {
		r.Engine.addRoute(next)
	} else {
		// 关键修复：即使路由已存在，也要重新计算 handler 链
		next.flatten()
	}
	return next
}

func (r *Router) handle(httpMethod, absolutePath string, handler HandlerCompose) *Router {
	next, find := r.Engine.getRouter(absolutePath)
	if _, composed := r.Engine.mixComposed(absolutePath); composed != nil {
		next.composed = compose(composed)
	}
	method := method{key: httpMethod, handler: handler}
	next.methods = append(next.methods, method)
	if !find {
		r.Engine.addRoute(next)
	} else {
		// 关键修复：即使路由已存在，也要重新计算 handler 链
		next.flatten()
	}
	return next
}

func (r *Router) Route(relativePath string, handles ...HandlerRoute) *Router {
	absolutePath := joinPaths(r.Path, relativePath)
	next, _ := r.Engine.getRouter(absolutePath)
	for _, handle := range handles {
		handle(next)
	}
	return next
}

func (r *Router) Use(relativePath string, handles ...Handler) *Router {
	absolutePath := joinPaths(r.Path, relativePath)
	next := r.use(absolutePath, makeCompose(handles...))
	return next
}

func (r *Router) ANY(handler Handler) *Router {
	return r.handle("ANY", r.Path, makeCompose(handler))
}

func (r *Router) Handle(httpMethod string, handler HandlerFinal) *Router {
	tempHandler := func(c *Context, next Next) {
		if err := handler(c); err != nil {
			if r.Engine.ErrorHandler != nil {
				r.Engine.ErrorHandler(c, err)
			} else {
				c.ErrorHTML(500, "Internal Server Error", err.Error())
			}
		}
	}
	return r.handle(httpMethod, r.Path, makeCompose(tempHandler))
}

func (r *Router) GET(handler HandlerFinal) *Router {
	return r.Handle("GET", handler)
}

func (r *Router) POST(handler HandlerFinal) *Router {
	return r.Handle("POST", handler)
}

func (r *Router) PUT(handler HandlerFinal) *Router {
	return r.Handle("PUT", handler)
}

func (r *Router) PATCH(handler HandlerFinal) *Router {
	return r.Handle("PATCH", handler)
}

func (r *Router) HEAD(handler HandlerFinal) *Router {
	return r.Handle("HEAD", handler)
}

func (r *Router) OPTIONS(handler HandlerFinal) *Router {
	return r.Handle("OPTIONS", handler)
}

func (r *Router) DELETE(handler HandlerFinal) *Router {
	return r.Handle("DELETE", handler)
}

func (r *Router) CONNECT(handler HandlerFinal) *Router {
	return r.Handle("CONNECT", handler)
}

func (r *Router) TRACE(handler HandlerFinal) *Router {
	return r.Handle("TRACE", handler)
}
