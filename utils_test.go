package cart

import (
	"testing"
)

// compose
func TestMakeCompose(t *testing.T) {
	temp := 0

	add2 := func(c *Context, next Next) {
		temp = temp + 2
		next()
	}
	plus2 := func(c *Context, next Next) {
		temp = temp * 2
		next()
	}
	add5 := func(c *Context, next Next) {
		temp = temp + 5
		next()
	}

	composed := makeCompose(add2, plus2, add5)
	composed(nil, func() {
		if temp != 9 {
			t.Errorf("Expected call compose func should return 9; got: %d", temp)
		}
	})()
}

func TestCompose(t *testing.T) {
	temp := 0
	add2 := func(c *Context, next Next) Next {
		return func() {
			temp = temp + 2
			next()
		}
	}
	plus2 := func(c *Context, next Next) Next {
		return func() {
			temp = temp * 2
			next()
		}
	}
	add5 := func(c *Context, next Next) Next {
		return func() {
			temp = temp + 5
			next()
		}
	}
	// (0+2)*2+5
	if compose() != nil {
		t.Errorf("Expected call compose func should return nil")
	}
	if compose(add2) == nil {
		t.Errorf("Expected call compose func should return add2")
	}
	composed := compose(add2, plus2, add5)(nil, func() {
		if temp != 9 {
			t.Errorf("Expected call compose func should return 9; got: %d", temp)
		}
	})
	composed()
}
