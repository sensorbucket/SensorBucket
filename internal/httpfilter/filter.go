package httpfilter

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrUnsupportedFieldType = errors.New("unsupported field type")
	ErrCantSetField         = errors.New("can't set field")
	ErrEmptyStrs            = errors.New("strs must have at least one item")
	ErrTMustBeStruct        = errors.New("T must be struct")
	ErrConvertingString     = errors.New("cant convert string")
)

// setSingleFieldValue sets a single value to the given field after converting the input string
// to the appropriate type based on the field's kind.
func setSingleFieldValue(field reflect.Value, str string) error {
	if !field.CanSet() {
		return ErrCantSetField
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(str)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		val, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: to int, %v", ErrConvertingString, err)
		}
		field.SetInt(val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		val, err := strconv.ParseUint(str, 10, 64)
		if err != nil {
			return fmt.Errorf("%w: to Uint, %v", ErrConvertingString, err)
		}
		field.SetUint(val)
	case reflect.Float32, reflect.Float64:
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return fmt.Errorf("%w: to Float, %v", ErrConvertingString, err)
		}
		field.SetFloat(val)
	case reflect.Bool:
		val, err := strconv.ParseBool(str)
		if err != nil {
			return fmt.Errorf("%w: to Bool, %v", ErrConvertingString, err)
		}
		field.SetBool(val)
	default:
		return ErrUnsupportedFieldType
	}

	return nil
}

// setFieldValue sets the value(s) from strs to the given field.
// If the field is a slice, it sets all values from strs, otherwise, it sets the first value.
func setFieldValue(field reflect.Value, strs []string) error {
	if !field.CanSet() {
		return ErrCantSetField
	}
	if len(strs) == 0 {
		return ErrEmptyStrs
	}
	if field.Kind() != reflect.Slice {
		return setSingleFieldValue(field, strs[0])
	}

	slice := reflect.MakeSlice(field.Type(), len(strs), len(strs))
	for i, str := range strs {
		element := slice.Index(i)
		err := setSingleFieldValue(element, str)
		if err != nil {
			return err
		}
	}
	field.Set(slice)
	return nil
}

type FieldConverter func(reflect.Value, []string) error
type FieldSingleConverter func(reflect.Value, string) error

// createSingleConverterFor creates a FieldSingleConverter for the given value kind.
// The returned function converts the given string and sets it to the value
func createSingleConverterFor(k reflect.Kind) (FieldSingleConverter, error) {
	switch k {
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
	k := t.Kind()
	if k != reflect.Slice {
		return singleToMultiConverter(createSingleConverterFor(k))
	}

	conv, err := createSingleConverterFor(t.Elem().Kind())
	if err != nil {
		return nil, err
	}
	return func(v reflect.Value, s []string) error {
		newSlice := reflect.MakeSlice(v.Type(), len(s), len(s))
		for ix, str := range s {
			el := newSlice.Index(ix)
			conv(el, str)
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
			if err != nil {
				return err
			}
		}
		return nil
	}

	return parse, nil
}
