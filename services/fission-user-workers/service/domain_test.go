package userworkers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/authtest"
	userworkers "sensorbucket.nl/sensorbucket/services/fission-user-workers/service"
)

func TestGetSetUserCode(t *testing.T) {
	usercode := []byte(`hello there this is test code`)
	w, err := userworkers.CreateWorker(authtest.DefaultTenantID, "test", "", usercode)
	require.NoError(t, err)
	usercode2, err := w.GetUserCode()
	require.NoError(t, err)
	assert.Equal(t, usercode, usercode2)
}
