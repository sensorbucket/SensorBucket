package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringToPermissions(t *testing.T) {
	testCases := []struct {
		desc                string
		strings             []string
		expectedError       error
		expectedPermissions Permissions
	}{
		{
			desc:                "valid permission",
			strings:             []string{string(WRITE_DEVICES)},
			expectedError:       nil,
			expectedPermissions: Permissions{WRITE_DEVICES},
		},
		{
			desc:                "invalid permission",
			strings:             []string{"INVALID_PERM"},
			expectedError:       ErrPermissionInvalid,
			expectedPermissions: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			perms, err := stringsToPermissions(tC.strings)
			assert.ErrorIs(t, err, tC.expectedError)
			assert.Equal(t, tC.expectedPermissions, perms)
		})
	}
}

func TestPermissionsFullfil(t *testing.T) {
	testCases := []struct {
		desc          string
		gotten        Permissions
		required      Permissions
		expectedError error
	}{
		{
			desc:          "Has no permissions with no requirements pass",
			gotten:        Permissions{},
			required:      Permissions{},
			expectedError: nil,
		},
		{
			desc:          "Has no permissions, but requires one should error",
			gotten:        Permissions{},
			required:      Permissions{READ_DEVICES},
			expectedError: ErrUnauthorized,
		},
		{
			desc:          "Has no permissions, but requires many should error",
			gotten:        Permissions{},
			required:      Permissions{READ_DEVICES, WRITE_API_KEYS},
			expectedError: ErrUnauthorized,
		},
		{
			desc:          "Has one permission, requires that permission should not errorb",
			gotten:        Permissions{READ_API_KEYS},
			required:      Permissions{READ_API_KEYS},
			expectedError: nil,
		},
		{
			desc:          "Has one permission, requires other permission should error",
			gotten:        Permissions{READ_DEVICES},
			required:      Permissions{READ_API_KEYS},
			expectedError: ErrUnauthorized,
		},
		{
			desc:          "Has partial overlap, should error",
			gotten:        Permissions{READ_DEVICES},
			required:      Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError: ErrUnauthorized,
		},
		{
			desc:          "Has many permissions, requires all those permissions",
			gotten:        Permissions{READ_DEVICES, WRITE_DEVICES},
			required:      Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := tC.gotten.Fulfills(tC.required)
			if tC.expectedError != nil {
				require.Error(t, err)
			} else {
				assert.ErrorIs(t, err, tC.expectedError)
			}
		})
	}
}
