package httpfilter

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrUnsupportedFieldType = errors.New("unsupported field type")
	ErrCantSetField         = errors.New("can't set field")
	ErrEmptyStrs            = errors.New("strs must have at least one item")
	ErrTMustBeStruct        = errors.New("T must be struct")
	ErrConvertingString     = errors.New("cant convert string")
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
				return fmt.Errorf("%w: %v", ErrConvertingString, err)
			}
			v.Set(reflect.ValueOf(val))
			return nil
		}, nil
	}

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
	if t.Kind() != reflect.Slice {
		return singleToMultiConverter(createSingleConverterFor(t))
	}

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

// Create creates a FilterCreator function for the given struct type T.
// It maps URL query keys to struct field indices and creates appropriate field converters.
func Create[T any]() (FilterCreator[T], error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	if t.Kind() != reflect.Struct {
		return nil, ErrTMustBeStruct
	}

	fieldCount := t.NumField()
	keys := make(map[int]string, fieldCount)
	converters := make(map[int]FieldConverter, fieldCount)

	for i := 0; i < fieldCount; i++ {
		ft := t.Field(i)
		if !ft.IsExported() {
			continue
		}
		tag, ok := ft.Tag.Lookup("url")
		if !ok {
			tag = strings.ToLower(ft.Name)
		}
		keys[i] = tag

		converter, err := createFieldConverter(ft.Type)
		if err != nil {
			return nil, err
		}
		converters[i] = converter
	}

	parse := func(q url.Values, val *T) error {
		v := reflect.ValueOf(val).Elem()
		for num, key := range keys {
			if !q.Has(key) {
				continue
			}
			field := v.Field(num)
			err := converters[num](field, q[key])
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
