package cart

import (
	"testing"
)

func TestRouterFlatten(t *testing.T) {
	e := New()

	m1 := func(c *Context, next Next) { next() }
	m2 := func(c *Context, next Next) { next() }
	m3 := func(c *Context, next Next) { next() }

	// Root middleware
	e.Use("/", m1)

	// Nested group with its own middleware
	v1 := e.Route("/v1", func(r *Router) {
		r.Use("/", m2)
		r.GET(func(c *Context) error {
			return nil
		})
	})

	// Deeply nested route
	v1.Route("/user", func(r *Router) {
		r.Use("/", m3)
		r.GET(func(c *Context) error {
			return nil
		})
	})

	// Check if handlers are pre-calculated correctly
	// /v1 should have m1, m2
	rV1, ok := e.findRouter("/v1")
	if !ok {
		t.Fatal("/v1 router not found")
	}
	if rV1.flattenHandlers["GET"] == nil {
		t.Error("/v1 GET handler not flattened")
	}

	// /v1/user should have m1, m2, m3
	rUser, ok := e.findRouter("/v1/user")
	if !ok {
		t.Fatal("/v1/user router not found")
	}
	if rUser.flattenHandlers["GET"] == nil {
		t.Error("/v1/user GET handler not flattened")
	}
}
