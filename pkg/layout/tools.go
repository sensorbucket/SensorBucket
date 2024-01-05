package layout

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
	Unknown SnackbarType = iota
	Success
	Error
)

func SnackbarDeleteSuccessful(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarSuccess(w, "Delete successful")
}

func SnackbarSaveSuccessful(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarSuccess(w, "Save successful")
}

func SnackbarBadRequest(w http.ResponseWriter, reason string) http.ResponseWriter {
	return WithSnackbarError(w, reason, http.StatusBadRequest)
}

func SnackbarSomethingWentWrong(w http.ResponseWriter) http.ResponseWriter {
	return WithSnackbarError(w, "Something went wrong", http.StatusInternalServerError)
}

func WithSnackbarError(w http.ResponseWriter, message string, statusCode int) http.ResponseWriter {
	return withSnackbarMessage(w, snackbarDetails{
		Message: message,
		Type:    Error,
	}, statusCode)
}

func WithSnackbarSuccess(w http.ResponseWriter, message string) http.ResponseWriter {
	return withSnackbarMessage(w, snackbarDetails{
		Message: message,
		Type:    Success,
	}, http.StatusOK)
}

type snackbarEvent struct {
	Details snackbarDetails `json:"showSnackbar"`
}

type snackbarDetails struct {
	Message string       `json:"message"`
	Type    SnackbarType `json:"type"`
}

func withSnackbarMessage(w http.ResponseWriter, details snackbarDetails, statusCode int) http.ResponseWriter {
	b, err := json.Marshal(snackbarEvent{
		Details: details,
	})
	if err != nil {
		log.Printf("[Warning] couldn't process snackbar message")
		b = snackbarGenericError()
	}
	w.Header().Set("hx-trigger", string(b))
	w.WriteHeader(statusCode)
	return w
}

func snackbarGenericError() []byte {
	ev := snackbarEvent{
		Details: snackbarDetails{
			Message: "Something went wrong",
			Type:    Error,
		},
	}
	b, _ := json.Marshal(ev)
	return b
}
