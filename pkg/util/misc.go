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
				if key == "*" || key == value {
					fValue = inValue.Field(i)
					fRet = ret.Field(i)
					fRet.Set(fValue)
				}
			}
		}
	}
	return nil
}

func GetMapWithoutPrefix(set map[string]string, prefix string) map[string]string {
	m := make(map[string]string)

	for key, value := range set {
		if strings.HasPrefix(key, prefix) {
			m[strings.TrimPrefix(key, prefix)] = value
		}
	}

	if len(m) == 0 {
		return nil
	}

	return m
}

// MoveSlice moves the element s[i] to index j in s.
func MoveSlice[S ~[]E, E any](s S, i, j int) {
	x := s[i]
	if i < j {
		copy(s[i:j], s[i+1:j+1])
	} else if i > j {
		copy(s[j+1:i+1], s[j:i])
	}
	s[j] = x
}
