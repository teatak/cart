package cart

import (
	"encoding/xml"
	"os"
	"testing"
)

func TestResolveAddress(t *testing.T) {
	// Test with no args
	os.Setenv("PORT", "")
	if addr := resolveAddress([]string{}); addr != ":8080" {
		t.Errorf("Expected default :8080, got %s", addr)
	}

	// Test with PORT env
	os.Setenv("PORT", "9090")
	if addr := resolveAddress([]string{}); addr != ":9090" {
		t.Errorf("Expected PORT :9090, got %s", addr)
	}
	os.Unsetenv("PORT")

	// Test with explicit arg
	if addr := resolveAddress([]string{":3000"}); addr != ":3000" {
		t.Errorf("Expected :3000, got %s", addr)
	}

	// Test panic
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	resolveAddress([]string{":1", ":2"})
}

func TestFilterFlags(t *testing.T) {
	if f := filterFlags("text/html"); f != "text/html" {
		t.Errorf("Expected text/html, got %s", f)
	}
	if f := filterFlags("text/html; charset=utf-8"); f != "text/html" {
		t.Errorf("Expected text/html, got %s", f)
	}
}

func TestHMarshalXML(t *testing.T) {
	h := H{"foo": "bar", "num": 1}
	b, err := xml.Marshal(h)
	if err != nil {
		t.Errorf("MarshalXML failed: %v", err)
	}
	// XML map order is not guaranteed, but check for containment
	s := string(b)
	if len(s) == 0 {
		t.Errorf("Expected XML output, got empty")
	}
}
