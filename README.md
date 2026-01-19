
```
 ██████╗ █████╗ ██████╗ ████████╗
██╔════╝██╔══██╗██╔══██╗╚══██╔══╝
██║     ███████║██████╔╝   ██║
██║     ██╔══██║██╔══██╗   ██║
╚██████╗██║  ██║██║  ██║   ██║
 ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝   ╚═╝
```

Current version: v2

A lightweight, expressive, and robust HTTP web framework for Go, inspired by Koa and Express.

## Features

- **Onion Architecture Middleware**: `func(ctx *Context, next Next)` style middleware.
- **Concise Routing**: Chainable routing API `c.Route("/a").Route("/b", ...)` with explicit tree structure.
- **Zero-Allocation Context Reuse**: Optimized for high performance.
- **Lifecycle Hooks**: `OnRequest` and `OnResponse` hooks for global intervention.
- **Graceful Shutdown**: Built-in support for safe server termination.
- **Centralized Error Handling**: Handlers can return errors directly.
- **Data Binding**: Easy JSON binding with validation support.

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

	// Global Middleware
	app.Use("/", func(c *cart.Context, next cart.Next) {
		fmt.Println("Before Request")
		next()
		fmt.Println("After Request")
	})

	// Lifecycle Hooks
	app.OnRequest = func(c *cart.Context) {
		fmt.Println("Request Started:", c.Request.URL.Path)
	}
	app.OnResponse = func(c *cart.Context) {
		fmt.Println("Request Finished")
	}

	// Routes
	app.Route("/hello").GET(func(c *cart.Context) error {
		return c.JSON(http.StatusOK, cart.H{
			"message": "Hello World",
		})
	})
	
	// Error Handling Example
	app.Route("/error").GET(func(c *cart.Context) error {
		return fmt.Errorf("something went wrong")
	})
	
	// Data Binding & Validation Example
	type User struct {
		Name  string `json:"name" binding:"required"` // Mandatory field
		Email string `json:"email"`
	}
	app.Route("/user").POST(func(c *cart.Context) error {
		var user User
		if err := c.BindJSON(&user); err != nil {
			return err // Will return error if 'name' is missing
		}
		return c.JSON(http.StatusOK, user)
	})

	// Static Files from embed.FS
	// app.Route("/static/*").Use(cart.StaticFS("/static", http.FS(myEmbedFS)))

	// Run with Graceful Shutdown
	app.RunGraceful(":8080")
}
```

## Core Concepts

### Middleware
Cart uses an "Onion" model for middleware, similar to Koa.

```go
app.Use("/", func(c *cart.Context, next cart.Next) {
    start := time.Now()
    next() // Pass control to the next middleware
    fmt.Printf("Latency: %v\n", time.Since(start))
})
```

### Routing
Routing is hierarchical and chainable.

```go
app.Route("/api").Route("/v1", func(r *cart.Router) {
    r.GET(func(c *cart.Context) error {
        c.String(200, "API V1 Root")
        return nil
    })
    
    r.Route("/users", func(sub *cart.Router) {
        sub.GET(func(c *cart.Context) error {
            // GET /api/v1/users
            return nil
        })
    })
})
```

### Data Validation
Supports `binding:"required"` tag. If validation fails, `Bind` methods will return an error.

### Error Handling
Handlers return `error`. You can define a global `ErrorHandler` in the engine.

```go
app.ErrorHandler = func(c *cart.Context, err error) {
    c.JSON(500, cart.H{"error": err.Error()})
}
```

### Graceful Shutdown
Use `RunGraceful` to start the server. It listens for `SIGINT` and `SIGTERM` signals and shuts down the server gracefully, waiting for active connections to complete.

```go
if err := app.RunGraceful(":8080"); err != nil {
    log.Fatal(err)
}
```
