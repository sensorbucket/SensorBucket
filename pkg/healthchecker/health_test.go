package healthchecker

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLivenessEndpoint(t *testing.T) {
	type scenario struct {
		checks         map[string]Check
		expectedStatus int
		expectedOk     []string
		expectedNotOk  []string
	}

	scenarios := map[string]scenario{
		"only 1 out of 3 checks succeed": {
			checks: map[string]Check{
				"failing-1": func() (string, bool) { return "check", false },
				"failing-2": func() (string, bool) { return "check", false },
				"success-1": func() (string, bool) { return "check", true },
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedOk:     []string{"success-1: check"},
			expectedNotOk:  []string{"failing-1: not check", "failing-2: not check"},
		},
		"all checks fail": {
			checks: map[string]Check{
				"failing-1": func() (string, bool) { return "check", false },
				"failing-2": func() (string, bool) { return "check", false },
				"failing-3": func() (string, bool) { return "check", false },
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedOk:     []string{},
			expectedNotOk:  []string{"failing-1: not check", "failing-2: not check", "failing-3: not check"},
		},
		"all checks succeed": {
			checks: map[string]Check{
				"success-1": func() (string, bool) { return "check", true },
				"success-2": func() (string, bool) { return "check", true },
				"success-3": func() (string, bool) { return "check", true },
				"success-4": func() (string, bool) { return "check", true },
			},
			expectedStatus: http.StatusOK,
			expectedOk:     []string{"success-1: check", "success-2: check", "success-3: check", "success-4: check"},
			expectedNotOk:  []string{},
		},
		"only 1 check fails": {
			checks: map[string]Check{
				"failing-1": func() (string, bool) { return "check", false },
				"success-1": func() (string, bool) { return "check", true },
				"success-2": func() (string, bool) { return "check", true },
				"success-3": func() (string, bool) { return "check", true },
				"success-4": func() (string, bool) { return "check", true },
			},
			expectedStatus: http.StatusServiceUnavailable,
			expectedOk:     []string{"success-1: check", "success-2: check", "success-3: check", "success-4: check"},
			expectedNotOk:  []string{"failing-1: not check"},
		},
		"no checks configured": {
			checks:         map[string]Check{},
			expectedStatus: http.StatusOK,
			expectedOk:     []string{},
			expectedNotOk:  []string{},
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			// Arrange
			builder := Create()
			for name, check := range cfg.checks {
				builder.AddLiveness(name, check)
			}
			handler, err := builder.BuildHandler()
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodGet, "/liveness", nil)
			w := httptest.NewRecorder()

			// Act
			handler.ServeHTTP(w, req)

			// Assert
			var result CheckResult
			err = json.NewDecoder(w.Body).Decode(&result)
			assert.NoError(t, err)
			assert.Equal(t, cfg.expectedStatus, w.Code)
			assert.ElementsMatch(t, cfg.expectedOk, result.Ok)
			assert.ElementsMatch(t, cfg.expectedNotOk, result.NotOk)
		})
	}
}

func TestReadinessEndpoint(t *testing.T) {
	// Arrange
	builder := Create()
	builder.AddReadiness("test", func() (string, bool) { return "ready", true })
	handler, err := builder.BuildHandler()
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/readiness", nil)
	w := httptest.NewRecorder()

	// Act
	handler.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAddLiveness(t *testing.T) {
	// Arrange
	builder := Create()

	// Act
	builder.AddLiveness("check1", func() (string, bool) { return "msg", false })
	builder.AddLiveness("check2", func() (string, bool) { return "msg", true })

	// Assert
	result := builder.liveness.Perform()
	assert.Contains(t, result.NotOk, "check1: not msg")
	assert.Contains(t, result.Ok, "check2: msg")
}

func TestAddReadiness(t *testing.T) {
	// Arrange
	builder := Create()

	// Act
	builder.AddReadiness("check1", func() (string, bool) { return "msg", false })
	builder.AddReadiness("check2", func() (string, bool) { return "msg", true })

	// Assert
	result := builder.readiness.Perform()
	assert.Contains(t, result.NotOk, "check1: not msg")
	assert.Contains(t, result.Ok, "check2: msg")
}

func TestWithEnv(t *testing.T) {
	// Arrange
	t.Setenv("HEALTH_ADDR", ":8082")

	// Act
	builder := Create().WithEnv()

	// Assert
	assert.Equal(t, ":8082", builder.Address)
}

func TestWithEnvMissing(t *testing.T) {
	// Arrange & Act
	builder := Create().WithEnv()

	// Assert
	assert.NotEmpty(t, builder.errors)
}

func TestBuildHandlerWithErrors(t *testing.T) {
	// Arrange
	builder := Create().WithEnv() // This will add an error since HEALTH_ADDR is not set

	// Act
	handler, err := builder.BuildHandler()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, handler)
	assert.NotEmpty(t, builder.errors)
}
