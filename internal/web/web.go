package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
)

// APIResponse ...
type APIResponse[T any] struct {
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}
type APIResponseAny = APIResponse[any]

// APIError is an API consumer friendly error
type APIError struct {
	Message    string `json:"message,omitempty"`
	Code       string `json:"code,omitempty"`
	HTTPStatus int    `json:"-"`
}

func (e *APIError) Error() string {
	return e.Message
}

func NewError(status int, message string, code string) *APIError {
	return &APIError{
		HTTPStatus: status,
		Message:    message,
		Code:       code,
	}
}

var (
	// ContentTypeError ...
	ContentTypeError = &APIError{
		HTTPStatus: 400,
		Message:    "Invalid content type",
		Code:       "INVALID_CONTENT_TYPE",
	}
	// InvalidJSONError ...
	InvalidJSONError = &APIError{
		HTTPStatus: 400,
		Message:    "Malformed JSON",
		Code:       "MALFORMED_JSON",
	}
)

// DecodeJSON ...
func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Header.Get("content-type") != "application/json" {
		return ContentTypeError
	}

	err := json.NewDecoder(r.Body).Decode(v)
	var apiError *APIError
	if errors.As(err, &apiError) {
		return apiError
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding JSON: %s\n", err)
		return InvalidJSONError
	}
	return nil
}

// HTTPResponse ...
func HTTPResponse(w http.ResponseWriter, s int, r interface{}) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(s)
	if err := json.NewEncoder(w).Encode(r); err != nil {
		log.Printf("HTTPResponse: failed to write response to client: %s\n", err.Error())
	}
}

// HTTPError writes the error to a http response writer
func HTTPError(w http.ResponseWriter, err error) {
	w.Header().Set("content-type", "application/json")

	var apierror *APIError
	if errors.As(err, &apierror) {
		w.WriteHeader(apierror.HTTPStatus)
		err := json.NewEncoder(w).Encode(&APIError{
			Message: err.Error(),
			Code:    apierror.Code,
		})
		if err != nil {
			log.Printf("HTTPError: failed to write response to client: %s\n", err.Error())
		}
		return
	}

	fmt.Printf("non APIError occured: %s\n", err)
	w.WriteHeader(500)
	err = json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Internal server error",
	})
	if err != nil {
		log.Printf("HTTPError: failed to write response to client: %s\n", err.Error())
	}
}
