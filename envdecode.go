// Package envdecode is a package for populating structs from environment
// variables, using struct tags.
package envdecode

import (
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// ErrInvalidTarget indicates that the target value passed to
// Decode is invalid.  Target must be a non-nil pointer to a struct.
var ErrInvalidTarget = errors.New("target must be non-nil pointer to struct")

// FailureFunc is called when an error is encountered during a MustDecode
// operation. It prints the error and terminates the process.
//
// This variable can be assigned to another function of the user-programmer's
// design, allowing for graceful recovery of the problem, such as loading
// from a backup configuration file.
var FailureFunc = func(err error) {
	log.Fatalf("envdecode: an error was encountered while decoding: %v\n", err)
}

// Decode environment variables into the provided target.  The target
// must be a non-nil pointer to a struct.  Fields in the struct must
// be exported, and tagged with an "env" struct tag with a value
// containing the name of the environment variable.  Default values
// may be provided by appending ",default=value" to the struct tag.
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

		parts := strings.Split(tag, ",")
		env := os.Getenv(parts[0])

		required := false
		hasDefault := false
		defaultValue := ""

		for _, o := range parts[1:] {
			if !required {
				required = strings.HasPrefix(o, "required")
			}
			if strings.HasPrefix(o, "default=") {
				hasDefault = true
				defaultValue = o[8:]
			}
		}

		if required && hasDefault {
			panic(`"default" and "required" may not be specified in the same annotation`)
		}
		if env == "" && required {
			return fmt.Errorf("the environment variable \"%s\" is missing", parts[0])
		}
		if env == "" {
			env = defaultValue
		}

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

// MustDecode calls Decode and terminates the process if any errors
// are encountered.
func MustDecode(target interface{}) {
	err := Decode(target)
	if err != nil {
		FailureFunc(err)
	}
}
