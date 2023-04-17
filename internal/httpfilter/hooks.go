package httpfilter

import (
	"encoding/json"
	"reflect"
	"time"
)

func sliceToSingleHook(from, to reflect.Type, i any) (any, error) {
	if from.Kind() == reflect.Slice && to.Kind() != reflect.Slice {
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

func stringToJSONRawMessage(from, to reflect.Type, data any) (any, error) {
	if to == reflect.TypeOf(json.RawMessage{}) && from == reflect.TypeOf([]string{}) {
		return json.RawMessage(data.([]string)[0]), nil
	}
	return data, nil
}
