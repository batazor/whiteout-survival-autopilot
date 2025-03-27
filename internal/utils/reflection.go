package utils

import (
	"fmt"
	"reflect"
	"strings"
)

func SetStateFieldByPath(obj interface{}, path string, value interface{}) error {
	fields := strings.Split(path, ".")
	v := reflect.ValueOf(obj)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return fmt.Errorf("object must be a non-nil pointer")
	}
	v = v.Elem()

	for i, field := range fields {
		if v.Kind() != reflect.Struct {
			return fmt.Errorf("field %q is not a struct", field)
		}
		v = v.FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(n, field)
		})
		if !v.IsValid() {
			return fmt.Errorf("field %q not found", field)
		}
		// if it's the last one → set it
		if i == len(fields)-1 {
			if !v.CanSet() {
				return fmt.Errorf("cannot set field %q", field)
			}
			val := reflect.ValueOf(value)
			if val.Type().ConvertibleTo(v.Type()) {
				v.Set(val.Convert(v.Type()))
				return nil
			}
			return fmt.Errorf("cannot convert %T to %s", value, v.Type())
		}
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
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

	for _, field := range fields {
		if v.Kind() != reflect.Struct {
			return nil, fmt.Errorf("field %q is not a struct", field)
		}
		v = v.FieldByNameFunc(func(n string) bool {
			return strings.EqualFold(n, field)
		})
		if !v.IsValid() {
			return nil, fmt.Errorf("field %q not found", field)
		}
		if v.Kind() == reflect.Ptr && !v.IsNil() {
			v = v.Elem()
		}
	}

	return v.Interface(), nil
}
