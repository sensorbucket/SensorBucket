package httpfilter

import (
	"reflect"
	"time"
)

func sliceToSingleHook(t1, t2 reflect.Type, i any) (any, error) {
	if t1.Kind() == reflect.Slice && t2.Kind() != reflect.Slice {
		return reflect.ValueOf(i).Index(0).Interface(), nil
	}
	return i, nil
}

func stringToTimeHook(from, to reflect.Type, data any) (any, error) {
	if to == reflect.TypeOf(time.Time{}) && from == reflect.TypeOf("") {
		return time.Parse(time.RFC3339, data.(string))
	}
	return data, nil
}
