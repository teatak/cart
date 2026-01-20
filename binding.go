package cart

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// bindingCache stores the meta-data of structs to avoid repeated reflection checks.
// key: reflect.Type, value: *structInfo
var bindingCache sync.Map

type structInfo struct {
	fields      []*fieldInfo
	fieldMap    map[string]*fieldInfo
	hasRequired bool
}

type fieldInfo struct {
	index    int
	name     string // parameter name from "form" tag or field struct name
	required bool   // if "binding" tag contains "required"
	kind     reflect.Kind
	elemKind reflect.Kind // for slices
	setter   func(reflect.Value, []string) error
}

func getStructInfo(typ reflect.Type) *structInfo {
	if val, ok := bindingCache.Load(typ); ok {
		return val.(*structInfo)
	}

	info := &structInfo{
		fields:   make([]*fieldInfo, 0, typ.NumField()),
		fieldMap: make(map[string]*fieldInfo, typ.NumField()),
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
		if required {
			info.hasRequired = true
		}

		fInfo := &fieldInfo{
			index:    i,
			name:     name,
			required: required,
			kind:     field.Type.Kind(),
		}

		if fInfo.kind == reflect.Slice {
			fInfo.elemKind = field.Type.Elem().Kind()
		}

		fInfo.setter = buildSetter(fInfo.kind, fInfo.elemKind, fInfo.name)

		info.fields = append(info.fields, fInfo)
		info.fieldMap[name] = fInfo
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
	if !info.hasRequired {
		return nil
	}

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
	if ptr == nil {
		return fmt.Errorf("binding element must be a non-nil pointer to struct")
	}
	typ := reflect.TypeOf(ptr)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer to struct")
	}
	val := reflect.ValueOf(ptr)
	if val.IsNil() {
		return fmt.Errorf("binding element must be a non-nil pointer to struct")
	}
	typ = typ.Elem()
	val = val.Elem()

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("binding element must be a struct")
	}

	info := getStructInfo(typ)

	for name, inputValue := range form {
		field, ok := info.fieldMap[name]
		if !ok {
			continue
		}
		if len(inputValue) == 0 {
			continue
		}
		structField := val.Field(field.index)
		if !structField.CanSet() {
			continue
		}

		if field.setter == nil {
			return fmt.Errorf("unsupported type %s for field %s", field.kind, field.name)
		}
		if err := field.setter(structField, inputValue); err != nil {
			return err
		}
	}
	return nil
}

func mapValuesQuery(ptr interface{}, rawQuery string) error {
	if rawQuery == "" {
		return nil
	}
	if ptr == nil {
		return fmt.Errorf("binding element must be a non-nil pointer to struct")
	}
	typ := reflect.TypeOf(ptr)
	if typ.Kind() != reflect.Ptr {
		return fmt.Errorf("binding element must be a pointer to struct")
	}
	val := reflect.ValueOf(ptr)
	if val.IsNil() {
		return fmt.Errorf("binding element must be a non-nil pointer to struct")
	}
	typ = typ.Elem()
	val = val.Elem()

	if typ.Kind() != reflect.Struct {
		return fmt.Errorf("binding element must be a struct")
	}

	info := getStructInfo(typ)
	seen := make([]bool, typ.NumField())

	start := 0
	for i := 0; i <= len(rawQuery); i++ {
		if i == len(rawQuery) || rawQuery[i] == '&' || rawQuery[i] == ';' {
			if i == start {
				start = i + 1
				continue
			}
			part := rawQuery[start:i]
			start = i + 1

			key := part
			valStr := ""
			if eq := strings.IndexByte(part, '='); eq >= 0 {
				key = part[:eq]
				valStr = part[eq+1:]
			}

			if key == "" {
				continue
			}
			keyDecoded := key
			if strings.IndexByte(key, '%') >= 0 || strings.IndexByte(key, '+') >= 0 {
				var err error
				keyDecoded, err = url.QueryUnescape(key)
				if err != nil {
					return err
				}
			}

			field, ok := info.fieldMap[keyDecoded]
			if !ok {
				continue
			}

			valDecoded := valStr
			if strings.IndexByte(valStr, '%') >= 0 || strings.IndexByte(valStr, '+') >= 0 {
				var err error
				valDecoded, err = url.QueryUnescape(valStr)
				if err != nil {
					return err
				}
			}

			structField := val.Field(field.index)
			if !structField.CanSet() {
				continue
			}

			if field.kind == reflect.Slice && field.elemKind == reflect.String {
				structField.Set(reflect.Append(structField, reflect.ValueOf(valDecoded)))
				continue
			}

			if seen[field.index] {
				continue
			}
			seen[field.index] = true

			if field.setter == nil {
				return fmt.Errorf("unsupported type %s for field %s", field.kind, field.name)
			}
			if err := field.setter(structField, []string{valDecoded}); err != nil {
				return err
			}
		}
	}
	return nil
}

func buildSetter(kind reflect.Kind, elemKind reflect.Kind, name string) func(reflect.Value, []string) error {
	switch kind {
	case reflect.Int:
		return func(field reflect.Value, input []string) error {
			return setIntField(input[0], 0, field)
		}
	case reflect.Int64:
		return func(field reflect.Value, input []string) error {
			return setIntField(input[0], 64, field)
		}
	case reflect.Bool:
		return func(field reflect.Value, input []string) error {
			return setBoolField(input[0], field)
		}
	case reflect.String:
		return func(field reflect.Value, input []string) error {
			field.SetString(input[0])
			return nil
		}
	case reflect.Slice:
		if elemKind == reflect.String {
			return func(field reflect.Value, input []string) error {
				field.Set(reflect.ValueOf(input))
				return nil
			}
		}
		return func(reflect.Value, []string) error {
			return fmt.Errorf("unsupported type %s for field %s", kind, name)
		}
	default:
		return func(reflect.Value, []string) error {
			return fmt.Errorf("unsupported type %s for field %s", kind, name)
		}
	}
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
