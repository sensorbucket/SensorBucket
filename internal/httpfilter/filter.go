package httpfilter

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/google/uuid"
	"github.com/gorilla/schema"
)

var d = schema.NewDecoder()

func init() {
	d.RegisterConverter(json.RawMessage{}, convertJSON)
	d.RegisterConverter(uuid.UUID{}, convertUUID)
}

func convertJSON(v string) reflect.Value {
	return reflect.ValueOf(json.RawMessage(v))
}

func convertUUID(v string) reflect.Value {
	id, err := uuid.Parse(v)
	if err != nil {
		return reflect.ValueOf(nil)
	}
	return reflect.ValueOf(id)
}

func Parse[T any](r *http.Request) (T, error) {
	var t T
	if err := d.Decode(&t, r.URL.Query()); err != nil {
		return t, err
	}
	return t, nil
}
