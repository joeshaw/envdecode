// envdecode is a package for populating structs from environment
// variables, using struct tags.
package envdecode

import (
	"errors"
	"os"
	"reflect"
	"strconv"
)

// ErrInvalidTarget indicates that the target value passed to
// Decode is invalid.  Target must be a non-nil pointer to a struct.
var ErrInvalidTarget = errors.New("target must be non-nil pointer to struct")

// Decode environment variables into the provided target.  The target
// must be a non-nil pointer to a struct.  Fields in the struct must
// be exported, and tagged with an "env" struct tag with a value
// containing the name of the environment variable.
func Decode(target interface{}) error {
	s := reflect.ValueOf(target)
	if s.Kind() != reflect.Ptr || s.IsNil() {
		return ErrInvalidTarget
	}

	s = s.Elem()
	if s.Kind() != reflect.Struct {
		return ErrInvalidTarget
	}

	t := s.Type()
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)

		switch f.Kind() {
		case reflect.Ptr:
			if f.Elem().Kind() != reflect.Struct {
				break
			}

			f = f.Elem()
			fallthrough

		case reflect.Struct:
			ss := f.Addr().Interface()
			Decode(ss)
		}

		if !f.CanSet() {
			continue
		}

		tag := t.Field(i).Tag.Get("env")
		if tag == "" {
			continue
		}

		env := os.Getenv(tag)
		if env == "" {
			continue
		}

		switch f.Kind() {
		case reflect.Bool:
			v, err := strconv.ParseBool(env)
			if err == nil {
				f.SetBool(v)
			}

		case reflect.Float32, reflect.Float64:
			bits := f.Type().Bits()
			v, err := strconv.ParseFloat(env, bits)
			if err == nil {
				f.SetFloat(v)
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			bits := f.Type().Bits()
			v, err := strconv.ParseInt(env, 0, bits)
			if err == nil {
				f.SetInt(v)
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			bits := f.Type().Bits()
			v, err := strconv.ParseUint(env, 0, bits)
			if err == nil {
				f.SetUint(v)
			}

		case reflect.String:
			f.SetString(env)
		}
	}

	return nil
}
