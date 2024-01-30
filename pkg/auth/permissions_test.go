package auth

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPermissionsValid(t *testing.T) {
	type testCase struct {
		permission  permission
		expectedErr error
	}
	scenarios := map[string]testCase{
		"valid permission": {
			permission:  permission("WRITE_USER_WORKERS"),
			expectedErr: nil,
		},
		"invalid permission": {
			permission:  permission("WEIRD_PERMISSION"),
			expectedErr: fmt.Errorf("WEIRD_PERMISSION is not a valid permission"),
		},
	}
	for scene, tc := range scenarios {
		t.Run(scene, func(t *testing.T) {
			assert.Equal(t, tc.expectedErr, tc.permission.Valid())
		})
	}
}

func TestAllAllowedPermissions(t *testing.T) {
	assert.Equal(t, []permission{
		READ_DEVICES,
		WRITE_DEVICES,
		READ_API_KEYS,
		WRITE_API_KEYS,
		READ_TENANTS,
		WRITE_TENANTS,
		READ_MEASUREMENTS,
		WRITE_MEASUREMENTS,
		READ_TRACING,
		READ_USER_WORKERS,
		WRITE_USER_WORKERS,
	}, AllAllowedPermissions())
}
