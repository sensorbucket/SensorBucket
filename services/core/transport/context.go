package coretransport

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"sensorbucket.nl/sensorbucket/internal/web"
)

type middleware = func(http.Handler) http.Handler

type ctxKey string

var ctxDeviceKey ctxKey = "device"

func (t *CoreTransport) useDeviceResolver() middleware {
	return func(next http.Handler) http.Handler {
		mw := func(rw http.ResponseWriter, r *http.Request) {
			idString := chi.URLParam(r, "device_id")
			id, err := strconv.ParseInt(idString, 10, 64)
			if err != nil {
				web.HTTPError(rw, ErrHTTPDeviceIDInvalid)
				return
			}

			dev, err := t.deviceService.GetDevice(r.Context(), id)
			if err != nil {
				web.HTTPError(rw, err)
				return
			}

			r = r.WithContext(context.WithValue(
				r.Context(),
				ctxDeviceKey,
				dev,
			))

			next.ServeHTTP(rw, r)
		}
		return http.HandlerFunc(mw)
	}
}

func urlParamInt64(r *http.Request, name string) (int64, error) {
	q := strings.Trim(chi.URLParam(r, name), " \r\n")
	if q == "" {
		return 0, web.NewError(http.StatusBadRequest, fmt.Sprintf("could not parse url parameter: missing %s url parameter", name), "")
	}
	i, err := strconv.ParseInt(q, 10, 64)
	if err != nil {
		return 0, web.NewError(http.StatusBadRequest, fmt.Sprintf("parameter %s is not an integer: %s", name, err), "")
	}
	return i, nil
}
