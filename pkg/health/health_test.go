package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestReadyEndpoint(t *testing.T) {
	type scenario struct {
		checks         map[string]Check
		expected       result
		expectedStatus int
	}

	// Arrange
	scenarios := map[string]scenario{
		"only 1 out of 3 checks succeed": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "failing-2": func() bool { return false }, "success-1": func() bool { return true }},
			expected: result{
				Message: "1/3 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{"success-1"},
					ChecksFailed: []string{"failing-1", "failing-2"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"all checks fail": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "failing-2": func() bool { return false }, "failing-3": func() bool { return false }},
			expected: result{
				Message: "0/3 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{},
					ChecksFailed: []string{"failing-1", "failing-2", "failing-3"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"all checks succeed": {
			checks: map[string]Check{"success-1": func() bool { return true }, "success-2": func() bool { return true }, "success-3": func() bool { return true }, "success-4": func() bool { return true }},
			expected: result{
				Message: "4/4 checks passed",
				Data: resultData{
					Success:      true,
					ChecksSucess: []string{"success-1", "success-2", "success-3", "success-4"},
					ChecksFailed: []string{},
				},
			},
			expectedStatus: http.StatusOK,
		},
		"only 1 check fails": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "success-1": func() bool { return true }, "success-2": func() bool { return true }, "success-3": func() bool { return true }, "success-4": func() bool { return true }},
			expected: result{
				Message: "4/5 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{"success-1", "success-2", "success-3", "success-4"},
					ChecksFailed: []string{"failing-1"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"no checks configured": {
			checks: map[string]Check{},
			expected: result{
				Message: "0/0 checks passed",
				Data:    resultData{},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			transport := testHealthEndpoint(nil, cfg.checks)
			req, _ := http.NewRequest("GET", "/readyz", nil)

			// Act
			rr := httptest.NewRecorder()
			transport.router.ServeHTTP(rr, req)

			// Assert
			result := asResult(rr.Body.String())
			assert.Equal(t, cfg.expectedStatus, rr.Code)
			assert.Equal(t, cfg.expected.Message, result.Message)
			assert.Equal(t, cfg.expected.Data.Success, result.Data.Success)
			assert.True(t, sliceEqual(cfg.expected.Data.ChecksFailed, result.Data.ChecksFailed))
			assert.True(t, sliceEqual(cfg.expected.Data.ChecksSucess, result.Data.ChecksSucess))
		})
	}
}

func TestLivelinessEndpoint(t *testing.T) {
	type scenario struct {
		checks         map[string]Check
		expected       result
		expectedStatus int
	}

	// Arrange
	scenarios := map[string]scenario{
		"only 1 out of 3 checks succeed": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "failing-2": func() bool { return false }, "success-1": func() bool { return true }},
			expected: result{
				Message: "1/3 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{"success-1"},
					ChecksFailed: []string{"failing-1", "failing-2"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"all checks fail": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "failing-2": func() bool { return false }, "failing-3": func() bool { return false }},
			expected: result{
				Message: "0/3 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{},
					ChecksFailed: []string{"failing-1", "failing-2", "failing-3"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"all checks succeed": {
			checks: map[string]Check{"success-1": func() bool { return true }, "success-2": func() bool { return true }, "success-3": func() bool { return true }, "success-4": func() bool { return true }},
			expected: result{
				Message: "4/4 checks passed",
				Data: resultData{
					Success:      true,
					ChecksSucess: []string{"success-1", "success-2", "success-3", "success-4"},
					ChecksFailed: []string{},
				},
			},
			expectedStatus: http.StatusOK,
		},
		"only 1 check fails": {
			checks: map[string]Check{"failing-1": func() bool { return false }, "success-1": func() bool { return true }, "success-2": func() bool { return true }, "success-3": func() bool { return true }, "success-4": func() bool { return true }},
			expected: result{
				Message: "4/5 checks passed",
				Data: resultData{
					Success:      false,
					ChecksSucess: []string{"success-1", "success-2", "success-3", "success-4"},
					ChecksFailed: []string{"failing-1"},
				},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
		"no checks configured": {
			checks: map[string]Check{},
			expected: result{
				Message: "0/0 checks passed",
				Data:    resultData{},
			},
			expectedStatus: http.StatusServiceUnavailable,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			transport := testHealthEndpoint(cfg.checks, nil)
			req, _ := http.NewRequest("GET", "/livez", nil)

			// Act
			rr := httptest.NewRecorder()
			transport.router.ServeHTTP(rr, req)

			// Assert
			result := asResult(rr.Body.String())
			assert.Equal(t, cfg.expectedStatus, rr.Code)
			assert.Equal(t, cfg.expected.Message, result.Message)
			assert.Equal(t, cfg.expected.Data.Success, result.Data.Success)
			assert.True(t, sliceEqual(cfg.expected.Data.ChecksFailed, result.Data.ChecksFailed))
			assert.True(t, sliceEqual(cfg.expected.Data.ChecksSucess, result.Data.ChecksSucess))
		})
	}
}

func TestWithLiveChecks(t *testing.T) {
	// Arrange
	hc := testHealthEndpoint(map[string]Check{}, map[string]Check{})
	checks := map[string]Check{
		"check1": func() bool { return false },
		"check2": func() bool { return true },
	}

	// Act
	hc.WithLiveChecks(checks)

	// Assert
	assert.False(t, hc.livelinessChecks["check1"]())
	assert.True(t, hc.livelinessChecks["check2"]())
}

func TestWithReadyChecks(t *testing.T) {
	// Arrange
	hc := testHealthEndpoint(map[string]Check{}, map[string]Check{})
	checks := map[string]Check{
		"check1": func() bool { return false },
		"check2": func() bool { return true },
	}

	// Act
	hc.WithReadyChecks(checks)

	// Assert
	assert.False(t, hc.readinessChecks["check1"]())
	assert.True(t, hc.readinessChecks["check2"]())
}

type result struct {
	Message string     `json:"message"`
	Data    resultData `json:"data"`
}

type resultData struct {
	Success      bool     `json:"success"`
	ChecksSucess []string `json:"checks_success"`
	ChecksFailed []string `json:"checks_failed"`
}

func asResult(jsonStr string) result {
	r := result{}
	_ = json.NewDecoder(strings.NewReader(jsonStr)).Decode(&r)
	return r
}

func sliceEqual(sl1 []string, sl2 []string) bool {
	if len(sl1) != len(sl2) {
		return false
	}
	for _, x := range sl1 {
		found := false
		for _, y := range sl2 {
			if x == y {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func testHealthEndpoint(liveChecks, readyChecks map[string]Check) *HealthChecker {
	hc := HealthChecker{
		livelinessChecks: liveChecks,
		readinessChecks:  readyChecks,
		router:           chi.NewRouter(),
	}
	hc.setupRoutes(hc.router)
	return &hc
}
