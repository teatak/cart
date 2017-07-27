package cart

import "testing"

func Equal(t *testing.T,a interface{},b interface{},err string) {
	if(a != b) {
		t.Error(err)
	}
}