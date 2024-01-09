package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProtect(t *testing.T) {

	type testCase struct {
		values             map[string]interface{}
		expectedStatusCode int
		expectedNextCalls  int
	}
	scenarios := map[string]testCase{
		"all required values present": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       []permission{READ_API_KEYS},
				"user_id":           int64(124),
			},
			expectedStatusCode: 200,
			expectedNextCalls:  1,
		},
		"current_tenant_id is missing": {
			values: map[string]interface{}{
				"permissions": []permission{READ_API_KEYS},
				"user_id":     int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"permissions is missing": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"user_id":           int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"user_id is missing": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       []permission{READ_API_KEYS},
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"all required values are missing": {
			values:             map[string]interface{}{},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"current_tenant_id is wrong type": {
			values: map[string]interface{}{
				"current_tenant_id": "123", // should be []int64!
				"permissions":       []permission{READ_API_KEYS},
				"user_id":           int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"permissions is wrong type": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       54325,
				"user_id":           int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"user_id is wrong type": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       []permission{READ_API_KEYS},
				"user_id":           "asdasdsad",
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"current_tenant_id is nil": {
			values: map[string]interface{}{
				"current_tenant_id": nil,
				"permissions":       []permission{READ_API_KEYS},
				"user_id":           int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"permissions is nil": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       nil,
				"user_id":           int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"user_id is nil": {
			values: map[string]interface{}{
				"current_tenant_id": []int64{12, 54, 13},
				"permissions":       []permission{READ_API_KEYS},
				"user_id":           nil,
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			ctx := testAccumulateContext(context.Background(), cfg.values)

			next := HandlerMock{
				ServeHTTPFunc: func(responseWriter http.ResponseWriter, request *http.Request) {},
			}

			handler := Protect()
			s := http.ServeMux{}
			s.Handle("/", handler(&next))

			// Act
			s.ServeHTTP(rr, req.WithContext(ctx))

			// Assert
			assert.Equal(t, cfg.expectedStatusCode, rr.Result().StatusCode)
			assert.Len(t, next.ServeHTTPCalls(), cfg.expectedNextCalls)
		})
	}
}

func testAccumulateContext(ctx context.Context, values map[string]interface{}) context.Context {
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}
