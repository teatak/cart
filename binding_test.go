package cart

import (
	"reflect"
	"testing"
)

func TestGetStructInfo(t *testing.T) {
	type TestStruct struct {
		Name     string `form:"name" binding:"required"`
		Age      int    `form:"age"`
		Active   bool   `form:"active"`
		Tags     []string
		unexport string
	}

	info := getStructInfo(reflect.TypeOf(TestStruct{}))
	if len(info.fields) != 4 { // Name, Age, Active, Tags (unexport is skipped)
		t.Errorf("Expected 4 fields, got %d", len(info.fields))
	}

	// Check Name field
	var nameField *fieldInfo
	for _, f := range info.fields {
		if f.name == "name" {
			nameField = f
			break
		}
	}
	if nameField == nil {
		t.Fatal("Expected name field")
	}
	if !nameField.required {
		t.Error("Expected name to be required")
	}
}

func TestValidate(t *testing.T) {
	type Required struct {
		Name string `binding:"required"`
	}

	// Test missing required field
	err := validate(&Required{})
	if err == nil {
		t.Error("Expected error for missing required field")
	}

	// Test with required field
	err = validate(&Required{Name: "test"})
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestValidateNonStruct(t *testing.T) {
	// Validate non-struct should return nil
	str := "not a struct"
	err := validate(&str)
	if err != nil {
		t.Errorf("Expected nil for non-struct, got %v", err)
	}
}

func TestMapValues(t *testing.T) {
	type Form struct {
		Name   string   `form:"name"`
		Age    int      `form:"age"`
		Age64  int64    `form:"age64"`
		Active bool     `form:"active"`
		Tags   []string `form:"tags"`
	}

	form := map[string][]string{
		"name":   {"John"},
		"age":    {"30"},
		"age64":  {"9999999999"},
		"active": {"true"},
		"tags":   {"go", "web"},
	}

	var data Form
	err := mapValues(&data, form)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if data.Name != "John" {
		t.Errorf("Expected Name=John, got %s", data.Name)
	}
	if data.Age != 30 {
		t.Errorf("Expected Age=30, got %d", data.Age)
	}
	if data.Age64 != 9999999999 {
		t.Errorf("Expected Age64=9999999999, got %d", data.Age64)
	}
	if !data.Active {
		t.Error("Expected Active=true")
	}
	if len(data.Tags) != 2 || data.Tags[0] != "go" {
		t.Errorf("Expected Tags=[go, web], got %v", data.Tags)
	}
}

func TestMapValuesEmptyValues(t *testing.T) {
	type Form struct {
		Name string `form:"name"`
		Age  int    `form:"age"`
	}

	form := map[string][]string{
		"name": {},
		"age":  {""},
	}

	var data Form
	err := mapValues(&data, form)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if data.Age != 0 {
		t.Errorf("Expected Age=0 for empty value, got %d", data.Age)
	}
}

func TestSetIntField(t *testing.T) {
	var val int64
	field := reflect.ValueOf(&val).Elem()

	err := setIntField("42", 64, field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got %d", val)
	}

	// Empty string
	err = setIntField("", 64, field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected 0 for empty, got %d", val)
	}
}

func TestSetBoolField(t *testing.T) {
	var val bool
	field := reflect.ValueOf(&val).Elem()

	err := setBoolField("true", field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !val {
		t.Error("Expected true")
	}

	err = setBoolField("false", field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val {
		t.Error("Expected false")
	}

	// Empty string
	err = setBoolField("", field)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if val {
		t.Error("Expected false for empty")
	}
}

func TestIsZeroValue(t *testing.T) {
	// String
	if !isZeroValue(reflect.ValueOf("")) {
		t.Error("Expected empty string to be zero")
	}
	if isZeroValue(reflect.ValueOf("hello")) {
		t.Error("Expected non-empty string to be non-zero")
	}

	// Int
	if !isZeroValue(reflect.ValueOf(0)) {
		t.Error("Expected 0 to be zero")
	}
	if isZeroValue(reflect.ValueOf(1)) {
		t.Error("Expected 1 to be non-zero")
	}

	// Bool
	if !isZeroValue(reflect.ValueOf(false)) {
		t.Error("Expected false to be zero")
	}
	if isZeroValue(reflect.ValueOf(true)) {
		t.Error("Expected true to be non-zero")
	}

	// Slice
	if !isZeroValue(reflect.ValueOf([]string{})) {
		t.Error("Expected empty slice to be zero")
	}

	// Pointer
	var ptr *int
	if !isZeroValue(reflect.ValueOf(ptr)) {
		t.Error("Expected nil pointer to be zero")
	}

	// Float
	if !isZeroValue(reflect.ValueOf(0.0)) {
		t.Error("Expected 0.0 to be zero")
	}

	// Uint
	if !isZeroValue(reflect.ValueOf(uint(0))) {
		t.Error("Expected uint(0) to be zero")
	}
}
