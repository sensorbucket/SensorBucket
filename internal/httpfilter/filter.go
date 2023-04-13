package httpfilter

import (
	"net/http"

	"github.com/mitchellh/mapstructure"
)

func Parse[T any](r *http.Request) (T, error) {
	var t T
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			sliceToSingleHook, stringToTimeHook,
		),
		WeaklyTypedInput: true,
		Squash:           true,
		TagName:          "url",
		Result:           &t,
	})
	if err := decoder.Decode(r.URL.Query()); err != nil {
		return t, err
	}
	return t, nil
}
