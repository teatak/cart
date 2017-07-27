package cart

type Next func()
//normal handler
type Handler func(*Context, Next)
type HandlerFinal func(*Context)
type HandlerCompose func(*Context, Next) Next

/*
transfer Handler to HandlerCompose func
 */
func makeCompose(handles ...Handler) HandlerCompose {
	composeHandles := []HandlerCompose{}
	for _, handle := range handles {
		innerHandle := handle
		tempHandle := func(c *Context, next Next) Next {
			return func() {
				innerHandle(c,next)
			}
		}
		composeHandles = append(composeHandles, tempHandle)
	}
	return compose(composeHandles...)
}

/*
compose HandlerCompose
	temp := 0
	A := func(c *Context, next Next) Next {
		return func() {
			temp = temp + 2;
			next()
		}
	}
	B := func(c *Context, next Next) Next {
		return func() {
			temp = temp * 2;
			next()
		}
	}
	composed := compose(A,B,B)(nil, func(){
		//this is the end the temp value is (0+2)*2*2
	})
	composed()
 */
func compose(functions ...HandlerCompose) HandlerCompose {
	if len(functions) == 0 {
		return nil
	}
	if len(functions) == 1 {
		return functions[0]
	}

	return func(c *Context, next Next) Next {
		last := functions[len(functions)-1]
		rest := functions[0:len(functions)-1]
		composed := last(c, next);
		for i, _ := range rest {
			composed = rest[len(rest)-1-i](c, composed)
		}
		return composed
	}

}