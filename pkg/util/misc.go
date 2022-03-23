package util

import (
	"reflect"
	"strings"
)

// GetFieldNameByTag returns the field name of struct that match the given tag and value
func GetFieldNameByTag(s interface{}, tag, value string) string {
	rt := reflect.TypeOf(s)
	if rt.Kind() != reflect.Struct {
		panic("bad type")
	}
	for i := 0; i < rt.NumField(); i++ {
		f := rt.Field(i)
		v := strings.Split(f.Tag.Get(tag), ",")[0] // use split to ignore tag "options"
		if v == value {
			return f.Name
		}
		if f.Type.Kind() == reflect.Struct && f.Anonymous {
			if name := GetFieldNameByTag(reflect.New(f.Type).Elem().Interface(), tag, value); name != "" {
				return name
			}
		}
	}
	return ""
}

// PruneByTag returns a copy of "in" that only contains fields with the given tag and value
func PruneByTag(in interface{}, value string, tag string) (interface{}, error) {
	inValue := reflect.ValueOf(in)

	ret := reflect.New(inValue.Type()).Elem()

	if err := prune(inValue, ret, value, tag); err != nil {
		return nil, err
	}
	return ret.Interface(), nil
}

func prune(inValue reflect.Value, ret reflect.Value, value string, tag string) error {
	switch inValue.Kind() {
	case reflect.Ptr:
		if inValue.IsNil() {
			return nil
		}
		if ret.IsNil() {
			// init ret and go to next level
			ret.Set(reflect.New(inValue.Type().Elem()))
		}
		return prune(inValue.Elem(), ret.Elem(), value, tag)
	case reflect.Struct:
		var fValue reflect.Value
		var fRet reflect.Value
		// search tag that has key equal to value
		for i := 0; i < inValue.NumField(); i++ {
			f := inValue.Type().Field(i)
			if key, ok := f.Tag.Lookup(tag); ok {
				if key == value {
					fValue = inValue.Field(i)
					fRet = ret.Field(i)
					fRet.Set(fValue)
				}
			}
		}
	}
	return nil
}
