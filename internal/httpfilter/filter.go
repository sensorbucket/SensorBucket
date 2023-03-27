package httpfilter

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrUnsupportedFieldType = errors.New("unsupported field type")
	ErrCantSetField         = errors.New("can't set field")
	ErrEmptyStrs            = errors.New("strs must have at least one item")
	ErrTMustBeStruct        = errors.New("T must be struct")
	ErrConvertingString     = errors.New("cant convert string")
	ErrMissingParameter     = web.NewError(http.StatusBadRequest, "missing required parameter", "ERR_QUERY_PARAMETER_MISSING")
	ErrBadParameterValue    = web.NewError(http.StatusBadRequest, "Invalid query parameter", "ERR_QUERY_PARAMETER_INVALID")
)

type StringConverter interface {
	FromString(string) (any, error)
}

var stringConverterType = reflect.TypeOf((*StringConverter)(nil)).Elem()

type FieldConverter func(reflect.Value, []string) error
type FieldSingleConverter func(reflect.Value, string) error

// createSingleConverterFor creates a FieldSingleConverter for the given value kind.
// The returned function converts the given string and sets it to the value
func createSingleConverterFor(t reflect.Type) (FieldSingleConverter, error) {

	if t.Implements(stringConverterType) {
		return func(v reflect.Value, s string) error {
			converter := reflect.New(t).Interface().(StringConverter)
			val, err := converter.FromString(s)
			if err != nil {
				return fmt.Errorf("%w: using StringConverter %v", ErrConvertingString, err)
			}
			v.Set(reflect.ValueOf(val))
			return nil
		}, nil
	}

	if t == reflect.TypeOf(time.Time{}) {
		return func(field reflect.Value, s string) error {
			t, err := time.Parse(time.RFC3339, s)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(t))
			return nil
		}, nil
	}

	// Fallback to more primitive types if no converter was returned yet
	switch t.Kind() {
	case reflect.String:
		return func(field reflect.Value, s string) error {
			field.SetString(s)
			return nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(field reflect.Value, s string) error {
			val, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return fmt.Errorf("%w: to int, %v", ErrConvertingString, err)
			}
			field.SetInt(val)
			return nil
		}, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(field reflect.Value, s string) error {
			val, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				return fmt.Errorf("%w: to uint, %v", ErrConvertingString, err)
			}
			field.SetUint(val)
			return nil
		}, nil
	case reflect.Float32, reflect.Float64:
		return func(field reflect.Value, s string) error {
			val, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return fmt.Errorf("%w: to float, %v", ErrConvertingString, err)
			}
			field.SetFloat(val)
			return nil
		}, nil
	case reflect.Bool:
		return func(field reflect.Value, s string) error {
			val, err := strconv.ParseBool(s)
			if err != nil {
				return fmt.Errorf("%w: to bool, %v", ErrConvertingString, err)
			}
			field.SetBool(val)
			return nil
		}, nil
	default:
		return nil, ErrUnsupportedFieldType
	}

}

// singleToMultiConverter converts a FieldSingleConverter to a FieldConverter by using
// the first value in the input string slice for the FieldSingleConverter.
func singleToMultiConverter(conv FieldSingleConverter, err error) (FieldConverter, error) {
	if conv == nil {
		return nil, err
	}
	return func(v reflect.Value, s []string) error {
		return conv(v, s[0])
	}, nil
}

// createFieldConverter creates a FieldConverter for the given field type.
// If the field type is a slice, it creates a converter for its elements and wraps
// it to handle slices.
func createFieldConverter(t reflect.Type) (FieldConverter, error) {
	// Directly try to get a converter for the field type
	// if no error occured, great! return converter
	// if a ErrUnsupportedFieldType occured and the field type is a slice, then DONT return
	//    and continue parsing the individual elements of the slice
	// Any other error or not slice, then return error
	fc, err := singleToMultiConverter(createSingleConverterFor(t))
	if err == nil {
		return fc, nil
	}
	if t.Kind() != reflect.Slice && !errors.Is(err, ErrUnsupportedFieldType) {
		return nil, err
	}

	// Create a converter for the elements of the slice
	conv, err := createSingleConverterFor(t.Elem())
	if err != nil {
		return nil, err
	}
	return func(v reflect.Value, s []string) error {
		newSlice := reflect.MakeSlice(v.Type(), len(s), len(s))
		for ix, str := range s {
			el := newSlice.Index(ix)
			if err := conv(el, str); err != nil {
				return err
			}
		}
		v.Set(newSlice)
		return nil
	}, nil
}

type FilterCreator[T any] func(q url.Values, val *T) error

func parseURLTag(ft reflect.StructField) (string, bool) {
	var required bool
	var key = strings.ToLower(ft.Name)

	tag, ok := ft.Tag.Lookup("url")
	if !ok {
		return key, required
	}

	tagParts := strings.Split(tag, ",")
	if tagParts[0] != "" {
		key = tagParts[0]
	}
	if len(tagParts) > 1 && tagParts[1] == "required" {
		required = true
	}
	return key, required
}

// Create creates a FilterCreator function for the given struct type T.
// It maps URL query keys to struct field indices and creates appropriate field converters.
func Create[T any]() (FilterCreator[T], error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, ErrTMustBeStruct
	}

	fieldCount := t.NumField()
	keys := make(map[int]string, fieldCount)
	required := make(map[int]bool, fieldCount)
	converters := make(map[int]FieldConverter, fieldCount)

	for i := 0; i < fieldCount; i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		keys[i], required[i] = parseURLTag(ft)
		converter, err := createFieldConverter(ft.Type)
		if err != nil {
			return nil, err
		}
		converters[i] = converter
	}

	parse := func(q url.Values, val *T) error {
		var err error

		v := reflect.ValueOf(val).Elem()
		for num, key := range keys {
			if !q.Has(key) {
				if required[num] {
					return fmt.Errorf("%w: %v", ErrMissingParameter, key)
				}
				continue
			}
			// Unescape values
			values := make([]string, len(q[key]))
			for ix, str := range q[key] {
				values[ix], err = url.QueryUnescape(str)
				if err != nil {
					return err
				}
			}

			// find converter
			field := v.Field(num)
			err := converters[num](field, values)
			if errors.Is(err, ErrConvertingString) {
				return fmt.Errorf("%w: %s, %v", ErrBadParameterValue, key, err)
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	return parse, nil
}

func MustCreate[T any]() FilterCreator[T] {
	f, err := Create[T]()
	if err != nil {
		panic(err)
	}
	return f
}
