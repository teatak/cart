package cart

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// bindingCache stores the meta-data of structs to avoid repeated reflection checks.
// key: reflect.Type, value: *structInfo
var bindingCache sync.Map

type structInfo struct {
	fields []*fieldInfo
}

type fieldInfo struct {
	index    int
	name     string // parameter name from "form" tag or field struct name
	required bool   // if "binding" tag contains "required"
	kind     reflect.Kind
	elemKind reflect.Kind // for slices
}

func getStructInfo(typ reflect.Type) *structInfo {
	if val, ok := bindingCache.Load(typ); ok {
		return val.(*structInfo)
	}

	info := &structInfo{
		fields: make([]*fieldInfo, 0, typ.NumField()),
	}

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		// Skip unexported fields
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		tag := field.Tag.Get("form")
		name := tag
		if name == "" {
			name = field.Name
		}

		bindingTag := field.Tag.Get("binding")
		required := strings.Contains(bindingTag, "required")

		fInfo := &fieldInfo{
			index:    i,
			name:     name,
			required: required,
			kind:     field.Type.Kind(),
		}

		if fInfo.kind == reflect.Slice {
			fInfo.elemKind = field.Type.Elem().Kind()
		}

		info.fields = append(info.fields, fInfo)
	}

	bindingCache.Store(typ, info)
	return info
}

// validate uses common struct info to validate fields (e.g. required).
func validate(obj interface{}) error {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	typ := val.Type()
	info := getStructInfo(typ)

	for _, field := range info.fields {
		if field.required {
			fieldVal := val.Field(field.index)
			if isZeroValue(fieldVal) {
				return fmt.Errorf("field '%s' is required", typ.Field(field.index).Name)
			}
		}
	}
	return nil
}

// mapValues binds map data (like form or query params) to a struct using cached info.
func mapValues(ptr interface{}, form map[string][]string) error {
	typ := reflect.TypeOf(ptr).Elem()
	val := reflect.ValueOf(ptr).Elem()

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("binding element must be a struct")
	}

	info := getStructInfo(typ)

	for _, field := range info.fields {
		structField := val.Field(field.index)
		if !structField.CanSet() {
			continue
		}

		inputValue, exists := form[field.name]
		if !exists {
			continue
		}

		if len(inputValue) == 0 {
			continue
		}

		switch field.kind {
		case reflect.Int:
			if err := setIntField(inputValue[0], 0, structField); err != nil {
				return err
			}
		case reflect.Int64:
			if err := setIntField(inputValue[0], 64, structField); err != nil {
				return err
			}
		case reflect.Bool:
			if err := setBoolField(inputValue[0], structField); err != nil {
				return err
			}
		case reflect.String:
			structField.SetString(inputValue[0])
		case reflect.Slice:
			if field.elemKind == reflect.String {
				structField.Set(reflect.ValueOf(inputValue))
			}
		// TODO: Add support for other types like Float, uint, etc if needed.
		default:
			// Originally we returned an error for unsupported types, but strictly
			// we might want to just ignore them or log warning.
			// For consistency with previous implementation, we return error if we can't handle it
			// BUT only if we matched a field name.
			return fmt.Errorf("unsupported type %s for field %s", field.kind, field.name)
		}
	}
	return nil
}

func setIntField(value string, bitSize int, field reflect.Value) error {
	if value == "" {
		value = "0"
	}
	intVal, err := strconv.ParseInt(value, 10, bitSize)
	if err == nil {
		field.SetInt(intVal)
	}
	return err
}

func setBoolField(value string, field reflect.Value) error {
	if value == "" {
		value = "false"
	}
	boolVal, err := strconv.ParseBool(value)
	if err == nil {
		field.SetBool(boolVal)
	}
	return err
}
