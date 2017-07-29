
```
 ██████╗ █████╗ ██████╗ ████████╗
██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
██║     ███████║██████╔╝   ██║
██║     ██╔══██║██╔══██╗   ██║
╚██████╗██║  ██║██║  ██║   ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
```

[![Build Status](https://travis-ci.org/gimke/cart.svg?branch=master)](https://travis-ci.org/gimke/cart) [![codecov](https://codecov.io/gh/gimke/cart/branch/master/graph/badge.svg)](https://codecov.io/gh/gimke/cart) [![GoDoc](https://godoc.org/github.com/gimke/cart?status.svg)](https://godoc.org/github.com/gimke/cart)


# Examples

## Using middleware
```go
func main() {
	c.Use("/favicon.ico", cart.Favicon("./public/favicon.ico"))
	c.Use("/", func(context *cart.Context, next cart.Next) {
		fmt.Println("A")
		next()
		fmt.Println("A")

	})
	c.Use("/admin/*file", cart.Static("./public",true))

	c.Use("/admin/*file", func(context *cart.Context, next cart.Next) {
            //go on process if file not found in ./public
            next()
	})
}
```
## Using GET POST ... 
```go
func main() {
	c.Route("/a").Route("/b", func(r *cart.Router) {
		r.ANY(func(context *cart.Context, next cart.Next) {
			context.Response.WriteString("ANY")
			next()
		})
		r.GET(func(context *cart.Context) {
			context.Response.WriteString(" /a/b")
		})
	})
	c.Route("/a/c", func(r *cart.Router) {
    		r.ANY(func(context *cart.Context, next cart.Next) {
    			context.Response.WriteString("ANY")
    			next()
    		})
    		r.GET(func(context *cart.Context) {
    			context.Response.WriteString(" /a/c")
    		})
    	})
}
```
