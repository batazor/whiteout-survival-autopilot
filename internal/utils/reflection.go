package utils

import (
	"fmt"
	"reflect"
	"strings"
)

// SetStateFieldByPath sets a field in a nested struct using a dot-notated path.
func SetStateFieldByPath(target any, path string, value any) error {
	parts := strings.Split(path, ".")
	v := reflect.ValueOf(target)

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("target must be a non-nil pointer")
	}
	v = v.Elem()

	for i, part := range parts {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("field '%s' is not a struct", part)
		}

		found := false
		t := v.Type()

		for fi := 0; fi < t.NumField(); fi++ {
			field := t.Field(fi)
			fieldVal := v.Field(fi)

			// 👇 обработка YAML-тега
			yamlTag := field.Tag.Get("yaml")
			yamlTag = strings.Split(yamlTag, ",")[0] // remove omitempty etc

			if part == field.Name || (yamlTag != "" && yamlTag == part) {
				// Инициализировать если nil
				if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
					newVal := reflect.New(fieldVal.Type().Elem())
					fieldVal.Set(newVal)
					fieldVal = newVal
				}
				if fieldVal.Kind() == reflect.Ptr {
					fieldVal = fieldVal.Elem()
				}
				if i == len(parts)-1 {
					if !fieldVal.CanSet() {
						return fmt.Errorf("cannot set field '%s'", part)
					}
					val := reflect.ValueOf(value)
					if val.Type().ConvertibleTo(fieldVal.Type()) {
						fieldVal.Set(val.Convert(fieldVal.Type()))
					} else {
						return fmt.Errorf("cannot convert %T to %s", value, fieldVal.Type().String())
					}
				} else {
					if fieldVal.Kind() == reflect.Struct {
						v = fieldVal
					} else {
						return fmt.Errorf("intermediate field '%s' is not a struct", part)
					}
				}
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("field '%s' not found", part)
		}
	}

	return nil
}

func GetStateFieldByPath(obj interface{}, path string) (interface{}, error) {
	fields := strings.Split(path, ".")
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil, fmt.Errorf("object must be a non-nil pointer")
	}
	v = v.Elem()

	for _, part := range fields {
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %q is not a struct", part)
		}

		t := v.Type()
		found := false

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			yamlTag := field.Tag.Get("yaml")
			yamlTag = strings.Split(yamlTag, ",")[0] // remove `,omitempty` etc

			if part == field.Name || (yamlTag != "" && part == yamlTag) {
				v = v.Field(i)
				if v.Kind() == reflect.Ptr && !v.IsNil() {
					v = v.Elem()
				}
				found = true
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("field %q not found", part)
		}
	}

	return v.Interface(), nil
}
