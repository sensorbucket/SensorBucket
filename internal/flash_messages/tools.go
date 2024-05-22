package flash_messages

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"

	"sensorbucket.nl/sensorbucket/pkg/api"
)

var base *url.URL

func U(str string, args ...any) string {
	res := fmt.Sprintf(str, args...)
	if base == nil {
		return res
	}
	return base.JoinPath(res).Path
}

func IsHX(r *http.Request) bool {
	return r.Header.Get("hx-request") == "true"
}

func SetBase(site *url.URL) {
	base = site
}

type SnackbarType int

const (
	SnackbarUnknown SnackbarType = iota
	SnackbarSuccess
	SnackbarError
)

func AddHTMXTrigger(w http.ResponseWriter, triggerName string, details interface{}) http.ResponseWriter {
	// HX-Trigger: {"event1":"A message", "event2":"Another message"}
	headerValue := w.Header().Get("hx-trigger")
	if headerValue != "" {
		// There is already a trigger present, add the new trigger to the existing header
		asMap := map[string]interface{}{}
		err := json.Unmarshal([]byte(headerValue), &asMap)
		if err != nil {
			log.Printf("[Warning] hx-trigger can't be set. existing header value invalid, err: %s", err)
			return w
		}

		// Add the new event
		asMap[triggerName] = details

		// Now reset the header after adding the new event
		headerBytes, err := json.Marshal(&asMap)
		if err != nil {
			log.Printf("[Warning] hx-trigger can't be set. updated header value invalid, err: %s", err)
			return w
		}
		w.Header().Set("hx-trigger", string(headerBytes))
		return w
	}

	b, err := json.Marshal(&details)
	if err != nil {
		log.Printf("[Warning] hx-trigger can't be set, invalid details, err: %s", err)
		return w
	}
	headerValue += fmt.Sprintf(`{"%s": %s}`, triggerName, string(b))
	w.Header().Set("hx-trigger", headerValue)
	return w
}

func SnackbarDeleteSuccessful(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarSuccess(w, "Delete successful")
}

func SnackbarSaveSuccessful(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarSuccess(w, "Save successful")
}

func SnackbarBadRequest(w http.ResponseWriter, reason string) http.ResponseWriter {
	return WithSnackbarError(w, reason)
}

func SnackbarSomethingWentWrong(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarError(w, "Something went wrong")
}

func WithSnackbarError(w http.ResponseWriter, message string) http.ResponseWriter {
	return withSnackbarMessage(w, snackbarDetails{
		Message: message,
		Type:    SnackbarError,
	})
}

func WithSnackbarSuccess(w http.ResponseWriter, message string) http.ResponseWriter {
	return withSnackbarMessage(w, snackbarDetails{
		Message: message,
		Type:    SnackbarSuccess,
	})
}

func IsAPIError(err error) (api.ApiError, bool) {
	generic, ok := err.(*api.GenericOpenAPIError)
	if !ok {
		return api.ApiError{}, false
	}
	apiErr, ok := generic.Model().(api.ApiError)
	return apiErr, ok
}

type snackbarDetails struct {
	Title   string       `json:"title"`
	Message string       `json:"message"`
	Type    SnackbarType `json:"type"`
	UID     string       `json:"uid"`
}

func withSnackbarMessage(w http.ResponseWriter, details snackbarDetails) http.ResponseWriter {
	w = AddHTMXTrigger(w, "showSnackbar", details)
	return w
}
