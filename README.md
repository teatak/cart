
```
 ██████╗ █████╗ ██████╗ ████████╗
██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
██║     ███████║██████╔╝   ██║
██║     ██╔══██║██╔══██╗   ██║
╚██████╗██║  ██║██║  ██║   ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
```

Current version: v2.0.0

A lightweight, expressive, and robust HTTP web framework for Go, inspired by Koa and Express, optimized for high-concurrency ⚡.

## Features

- **Extreme Performance**: Registration-time middleware flattening and zero-allocation context pooling.
- **Onion Architecture Middleware**: Powerful `func(ctx *Context, next Next)` style.
- **Hierarchical Routing**: Chainable and explicit tree structure.
- **Lifecycle Hooks**: `OnRequest` and `OnResponse` hooks for global intervention.
- **Built-in Validation**: Smart binding with `binding:"required"` support.
- **Graceful Shutdown**: Ready for production with safe termination.
- **Modern Standards**: Native support for `embed.FS`, CORS, Gzip, and RequestID.

## Installation

```bash
go get github.com/teatak/cart/v2
```

## Quick Start

```go
package main

import (
	"fmt"
	"net/http"
	"github.com/teatak/cart/v2"
)

func main() {
	app := cart.New()

	// 1. Standard Middlewares
	app.Use("/", cart.Logger())
	app.Use("/", cart.Recovery())

	// 2. Lifecycle Hooks
	app.OnRequest = func(c *cart.Context) {
		fmt.Println("Request Started:", c.Request.URL.Path)
	}

	// 3. Routing with Validation
	type User struct {
		ID   int    `form:"id" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	app.Route("/user/:id").POST(func(c *cart.Context) error {
		var user User
		if err := c.Bind(&user); err != nil {
			return err // Returns 400 with error message if validation fails
		}
		id, _ := c.ParamInt("id")
		fmt.Printf("User ID from path: %d\n", id)
		return c.JSON(http.StatusOK, user)
	})

	// 4. Run with Graceful Shutdown
	app.RunGraceful(":8080")
}
```

## Performance Highlights ⚡

Cart is designed for maximum throughput:
- **Context & Params Pooling**: Uses `sync.Pool` to reuse `Context` and `Params` objects, drastically reducing GC pressure.
- **Middleware Flattening**: Middleware chains are pre-composed during route registration. The runtime overhead of method mixing and middleware lookup is **ZERO**.
- **Radix Tree Routing**: Efficient path matching based on a high-performance radix tree.

## Core Concepts

### Middleware
Cart uses an "Onion" model. Calling `next()` executes the next handler in the chain.

```go
// Custom logger middleware
app.Use("/", func(c *cart.Context, next cart.Next) {
    start := time.Now()
    next() 
    fmt.Printf("[%s] %s %v\n", c.Request.Method, c.Request.URL.Path, time.Since(start))
})
```

#### Standard Middlewares
- `cart.Logger()`: Colored terminal output for requests.
- `cart.Recovery()`: Recovers from panics and returns 500.
- `cart.Gzip()`: Transparent Gzip compression.
- `cart.RequestID()`: Injects `X-Request-ID` into headers.
- `cart.CORS()`: Easy CORS configuration.
- `cart.StaticFS()`: Serve files from `embed.FS` or `http.Dir`.

### Data Binding & Validation
The `Bind()` method automatically detects `Content-Type` (JSON/Form) and validates the struct using the `binding` tag.

```go
type CreatePost struct {
    Title string `json:"title" binding:"required"`
    Body  string `json:"body" binding:"required"`
}
```

### Error Handling
Handlers return an `error`. You can catch all errors globally using `app.ErrorHandler`.

```go
app.ErrorHandler = func(c *cart.Context, err error) {
    c.JSON(http.StatusBadRequest, cart.H{"error": err.Error()})
}
```

### Lifecycle Hooks
- `OnRequest(c *Context)`: Called immediately after a request enters the server.
- `OnResponse(c *Context)`: Called after the response is sent (via `defer`).

## Advanced Context API

- `c.Context()`: Access the standard `context.Context`.
- `c.ParamInt(key)`: Parse route parameters as integers.
- `c.Query(key)` / `c.PostForm(key)`: Quick access to parameters.
- `c.AbortWithStatus(code)`: Stop execution and return status.
- `c.JSONP(code, callback, obj)`: Render JSONP for legacy browser support.

