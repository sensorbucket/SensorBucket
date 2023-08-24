package httpfilter

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/gorilla/schema"
)

var d = schema.NewDecoder()

func init() {
	d.RegisterConverter(json.RawMessage{}, convertJSON)
}

func convertJSON(v string) reflect.Value {
	return reflect.ValueOf(json.RawMessage(v))
}

func Parse[T any](r *http.Request) (T, error) {
	var t T
	if err := d.Decode(&t, r.URL.Query()); err != nil {
		return t, err
	}
	return t, nil
}
