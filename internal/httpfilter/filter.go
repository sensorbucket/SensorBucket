package httpfilter

import (
	"encoding/json"
	"fmt"
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
	val, err := uuid.Parse(v)
	if err != nil {
		reflect.ValueOf(fmt.Errorf("invalid uuid '%s'", v))
	}
	return reflect.ValueOf(val)
}

func Parse[T any](r *http.Request) (T, error) {
	var t T
	if err := d.Decode(&t, r.URL.Query()); err != nil {
		return t, err
	}
	return t, nil
}
