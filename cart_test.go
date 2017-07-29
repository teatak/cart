package cart

import "testing"

func Equal(t *testing.T,a interface{},b interface{},err string) {
	if(a != b) {
		t.Error(err)
	}
}

func TestEngine_New(t *testing.T) {
	c := New()
	c.Route("/a").Route("/b", func(router IRouter) {
		//
	})
}