
```
 ██████╗ █████╗ ██████╗ ████████╗
██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
██║     ███████║██████╔╝   ██║
██║     ██╔══██║██╔══██╗   ██║
╚██████╗██║  ██║██║  ██║   ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
```

Current version: v2.1.8

A lightweight, expressive, and robust HTTP web framework for Go, inspired by Koa and Express, optimized for high-concurrency ⚡.

## Features

- **Extreme Performance**: Registration-time middleware flattening and zero-allocation context pooling.
- **Onion Architecture Middleware**: Powerful `func(ctx *Context, next Next)` style with full `Abort()` support.
- **Hierarchical Routing**: Chainable and explicit tree structure with pre-calculated handler chains.
- **Lifecycle Hooks**: `OnRequest` and `OnResponse` hooks for global intervention.
- **Security-First**: `TrustedProxies` support to prevent IP spoofing in `ClientIP()`.
- **Production-Ready**: Configurable HTTP server timeouts and graceful shutdown support.
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
	"time"
	"github.com/teatak/cart/v2"
)

func main() {
	app := cart.New()

	// 1. Server Configuration
	app.ReadTimeout = 30 * time.Second
	app.TrustedProxies = []string{"127.0.0.1", "192.168.1.1"}

	// 2. Standard Middlewares
	app.Use("/", cart.Logger())
	app.Use("/", cart.Recovery())

	// 3. Routing with Validation
	type User struct {
		ID   int    `form:"id" binding:"required"`
		Name string `json:"name" binding:"required"`
	}

	app.Route("/user/:id").POST(func(c *cart.Context) error {
		var user User
		if err := c.Bind(&user); err != nil {
			return err 
		}
		id, _ := c.ParamInt("id")
		fmt.Printf("User ID from path: %d, Client IP: %s\n", id, c.ClientIP())
		return c.JSON(http.StatusOK, user)
	})

	// 4. Run with Graceful Shutdown
	app.RunGraceful(":8080")
}
```

## Performance Highlights ⚡

Cart is designed for maximum throughput and minimal GC pressure:
- **Zero-Allocation Hot Path**: Context and parameters are pooled via `sync.Pool`. Request handlers avoid closure allocations during the request cycle.
- **Middleware Pre-calculation**: All middleware chains (including inherited ones) are flattened into a single slice during registration. Runtime overhead of middleware lookup is **ZERO**.
- **Struct Caching**: Reflection overhead in data binding (`Bind`, `Validate`) is minimized using a concurrent-safe `sync.Map` cache for struct metadata.
- **Gzip Pooling**: `gzip.Writer` instances are recycled using `sync.Pool` to significantly reduce memory allocations during compression.
- **Radix Tree Routing**: High-performance path matching with support for parameters and catch-alls.
- **Smart Recovery**: In Release mode, `Recovery` middleware skips expensive source code reading to maximize speed and security.

## Core Concepts

### Middleware Control
Cart uses an "Onion" model with explicit control:
- `next()`: Execute the next handler.
- `c.Abort()`: Stop the chain immediately.

```go
app.Use("/admin", func(c *cart.Context, next cart.Next) {
    if !isAdmin(c) {
        c.AbortWithStatus(403)
        return
    }
    next() 
})
```

### Security: Trusted Proxies
To prevent IP spoofing, `cart` only parses `X-Forwarded-For` or `X-Real-IP` if the request originates from a `TrustedProxy`.

```go
app.TrustedProxies = []string{"10.0.0.1"} // Your Nginx/LB IP
```

### Customizable Server
Unlike most frameworks that hide the `http.Server` configuration, `cart` exposes it directly on the `Engine`:

```go
app.ReadTimeout = 60 * time.Second
app.WriteTimeout = 60 * time.Second
app.IdleTimeout = 120 * time.Second
```

## Advanced Context API

- `c.ClientIP()`: Secure client IP detection.
- `c.ParamInt(key)`: Parse route parameters as integers.
- `c.Bind(obj)`: Auto-detect Content-Type (JSON/Form) and validate.
- `c.HTML(code, name, data)`: Render templates with custom delims.
- `c.JSONP(code, callback, obj)`: Render JSONP for legacy support.
- `c.AbortWithStatus(code)`: Stop execution and return status.

## Error Handling

Cart provides a unified error handling mechanism:

```go
app := cart.New()

// Global error handler
app.ErrorHandler = func(c *cart.Context, err error) {
    c.JSON(500, cart.H{"error": err.Error()})
}

// In route handlers, just return the error
app.Route("/api").GET(func(c *cart.Context) error {
    data, err := fetchData()
    if err != nil {
        return err // Handled by ErrorHandler
    }
    return c.JSON(200, data)
})
```

## License

MIT License
