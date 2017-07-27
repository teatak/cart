package cart

import (
	"testing"
)

// Params
func TestParams(t *testing.T) {
	ps := Params{
		Param{"param1", "value1"},
		Param{"param2", "value2"},
		Param{"param3", "value3"},
	}
	for i := range ps {
		if val, _ := ps.Get(ps[i].Key); val != ps[i].Value {
			t.Errorf("Wrong value for %s: Got %s; Want %s", ps[i].Key, val, ps[i].Value)
		}
	}
	if val, _ := ps.Get("noKey"); val != "" {
		t.Errorf("Expected empty string for not found key; got: %s", val)
	}
}

func TestContext(t *testing.T) {
	c := &Context{}
	c.reset(nil,nil)
}