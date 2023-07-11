package e2e_test

import (
	"os"
	"strings"
	"testing"
)

func thisIsAnE2ETest(t *testing.T) {
	value, ok := os.LookupEnv("SENSORBUCKET_TEST_E2E")
	if !ok {
		t.SkipNow()
		return
	}
	value = strings.ToLower(value)
	if value != "yes" && value != "true" {
		t.SkipNow()
		return
	}
}
