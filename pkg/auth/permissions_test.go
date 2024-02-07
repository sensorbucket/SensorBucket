package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPermissionsValid(t *testing.T) {
	testCases := []struct {
		desc          string
		permission    Permission
		expectedError error
	}{
		{
			desc:          "valid permission",
			permission:    WRITE_DEVICES,
			expectedError: nil,
		},
		{
			desc:          "invalid permission",
			permission:    Permission("NON_EXISTANT"),
			expectedError: ErrPermissionInvalid,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.ErrorIs(t, tC.permission.Valid(), tC.expectedError)
		})
	}
}

func TestPermissionsFlattenCorrectly(t *testing.T) {
	testCases := []struct {
		desc     string
		set      Permissions
		expected []Permission
	}{
		{
			desc:     "Empty to empty",
			set:      Permissions{},
			expected: []Permission{},
		},
		{
			desc:     "One to one",
			set:      Permissions{READ_DEVICES},
			expected: []Permission{READ_DEVICES},
		},
		{
			desc:     "Two to Two",
			set:      Permissions{READ_DEVICES, WRITE_DEVICES},
			expected: []Permission{READ_DEVICES, WRITE_DEVICES},
		},
		{
			desc:     "Group with one to one",
			set:      Permissions{Permissions{READ_DEVICES}},
			expected: []Permission{READ_DEVICES},
		},
		{
			desc:     "Group with many to many",
			set:      Permissions{Permissions{READ_DEVICES, WRITE_DEVICES}},
			expected: []Permission{READ_DEVICES, WRITE_DEVICES},
		},
		{
			desc:     "Group with groups",
			set:      Permissions{Permissions{Permissions{READ_DEVICES}, Permissions{WRITE_DEVICES}}},
			expected: []Permission{READ_DEVICES, WRITE_DEVICES},
		},
		{
			desc:     "Remove duplicates",
			set:      Permissions{READ_DEVICES, READ_DEVICES},
			expected: []Permission{READ_DEVICES},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.set.Permissions())
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
