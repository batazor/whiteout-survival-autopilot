package config

import (
	"reflect"
)

// flattenStruct flattens a struct to dot-notated keys
func flattenStruct(prefix string, val any, out map[string]interface{}) {
	v := reflect.ValueOf(val)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		key := field.Name
		yamlTag := field.Tag.Get("yaml")
		if yamlTag != "" && yamlTag != "-" {
			key = yamlTag
		}

		fVal := v.Field(i)
		if !fVal.CanInterface() {
			continue
		}

		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch fVal.Kind() {
		case reflect.Struct:
			flattenStruct(fullKey, fVal.Interface(), out)
		case reflect.Map:
			mapOut := map[string]interface{}{}
			for _, mapKey := range fVal.MapKeys() {
				mapVal := fVal.MapIndex(mapKey)
				strKey := mapKey.String()
				mapOut[strKey] = mapVal.Interface()
			}
			out[fullKey] = mapOut
		default:
			out[fullKey] = fVal.Interface()
		}
	}
}
