package auth

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMustHavePermissionsRequestedPermissions(t *testing.T) {
	// Arrange
	type testCase struct {
		permissionsInCtx []permission
		permissionsInput []permission
		expectedErr      error
	}

	scenarios := map[string]testCase{
		"no permissions present": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES},
			permissionsInCtx: []permission{}, // empty!
			expectedErr:      ErrPermissionsNotGranted,
		},
		"no permissions present in context": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES},
			permissionsInCtx: nil,
			expectedErr:      ErrNoPermissions,
		},
		"some requested permissions are missing": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS, READ_DEVICES},
			expectedErr:      ErrPermissionsNotGranted,
		},
		"only 1 requested permission is missing": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS},
			expectedErr:      ErrPermissionsNotGranted,
		},
		"only 1 requested permission is present": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS},
			expectedErr:      ErrPermissionsNotGranted,
		},
		"all requested permissions are present": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			expectedErr:      nil,
		},
		"all requested permissions are missing": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS},
			expectedErr:      ErrPermissionsNotGranted,
		},
		"more permissions are present than are requested": {
			permissionsInput: []permission{READ_API_KEYS, READ_DEVICES},
			permissionsInCtx: []permission{READ_API_KEYS, WRITE_API_KEYS, WRITE_DEVICES, READ_DEVICES},
			expectedErr:      nil,
		},
	}
	for testC, cfg := range scenarios {
		t.Run(testC, func(t *testing.T) {
			ctx := context.Background()
			if cfg.permissionsInCtx != nil {
				ctx = context.WithValue(ctx, ctxPermissions, cfg.permissionsInCtx)
			}

			// Act
			err := MustHavePermissions(ctx, cfg.permissionsInput[0], cfg.permissionsInput[1:]...)

			// Assert
			assert.ErrorIs(t, err, cfg.expectedErr)
		})
	}
}

func TestHasRole(t *testing.T) {
	// Arrange
	type testCase struct {
		permissionsInCtx []permission
		roleInput        role
		expectedRes      bool
	}
	scenarios := map[string]testCase{
		"does not have requested role": {
			permissionsInCtx: []permission{READ_DEVICES, WRITE_DEVICES},
			roleInput:        role([]permission{READ_API_KEYS, WRITE_API_KEYS}),
			expectedRes:      false,
		},
		"has only the requested role": {
			permissionsInCtx: []permission{READ_API_KEYS, WRITE_API_KEYS},
			roleInput:        role([]permission{READ_API_KEYS, WRITE_API_KEYS}),
			expectedRes:      true,
		},
		"has requested role and more permissions": {
			permissionsInCtx: []permission{READ_API_KEYS, READ_DEVICES, WRITE_API_KEYS, WRITE_DEVICES},
			roleInput:        role([]permission{READ_API_KEYS, WRITE_API_KEYS}),
			expectedRes:      true,
		},
		"has no permissions": {
			permissionsInCtx: []permission{},
			roleInput:        role([]permission{READ_API_KEYS, WRITE_API_KEYS}),
			expectedRes:      false,
		},
		"has no permissions in context": {
			permissionsInCtx: nil,
			roleInput:        role([]permission{READ_API_KEYS, WRITE_API_KEYS}),
			expectedRes:      false,
		},
		"has only 1 role in context": {
			permissionsInCtx: []permission{},
			roleInput:        role([]permission{READ_API_KEYS}),
			expectedRes:      false,
		},
	}
	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			// Act
			ctx := context.Background()
			if cfg.permissionsInCtx != nil {
				ctx = context.WithValue(ctx, ctxPermissions, cfg.permissionsInCtx)
			}
			result := HasRole(ctx, cfg.roleInput)

			// Assert
			assert.Equal(t, cfg.expectedRes, result)
		})
	}
}

func TestGetTenants(t *testing.T) {
	// Arrange
	type testCase struct {
		tenantsInContext []int64
		expectedRes      []int64
		expectedErr      error
	}

	scenarios := map[string]testCase{
		"no tenants in context": {
			tenantsInContext: nil,
			expectedRes:      nil,
			expectedErr:      ErrNoTenantIDFound,
		},
		"no tenants": {
			tenantsInContext: []int64{},
			expectedRes:      []int64{},
			expectedErr:      nil,
		},
		"multiple tenants in context": {
			tenantsInContext: []int64{541, 241, 21},
			expectedRes:      []int64{541, 241, 21},
			expectedErr:      nil,
		},
		"only 1 tenant in context": {
			tenantsInContext: []int64{143},
			expectedRes:      []int64{143},
			expectedErr:      nil,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			ctx := context.Background()
			if cfg.tenantsInContext != nil {
				ctx = context.WithValue(ctx, ctxCurrentTenantID, cfg.tenantsInContext)
			}

			// Act
			result, err := GetTenants(ctx)

			// Assert
			assert.Equal(t, cfg.expectedRes, result)
			assert.Equal(t, cfg.expectedErr, err)
		})
	}
}

func TestHasPermissionsFor(t *testing.T) {
	// Arrange
	type testCase struct {
		tenantsInContext []int64
		tenantsInput     []int64
		expectedRes      bool
	}

	scenarios := map[string]testCase{
		"no tenants in context": {
			tenantsInContext: nil,
			tenantsInput:     []int64{123, 54, 21, 53},
			expectedRes:      false,
		},
		"no tenants": {
			tenantsInContext: []int64{},
			tenantsInput:     []int64{123, 54, 21, 53},
			expectedRes:      false,
		},
		"has permissions for 1 tenant": {
			tenantsInContext: []int64{123},
			tenantsInput:     []int64{123, 54, 21, 53},
			expectedRes:      false,
		},
		"has permissions for some tenants": {
			tenantsInContext: []int64{123, 54},
			tenantsInput:     []int64{123, 54, 21, 53},
			expectedRes:      false,
		},
		"has permissions for all tenants": {
			tenantsInContext: []int64{123, 54, 21, 53},
			tenantsInput:     []int64{123, 54, 21, 53},
			expectedRes:      true,
		},
		"has permissions for 1 tenant and 1 tenant is requested": {
			tenantsInContext: []int64{123},
			tenantsInput:     []int64{123},
			expectedRes:      true,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			ctx := context.Background()
			if cfg.tenantsInContext != nil {
				ctx = context.WithValue(ctx, ctxCurrentTenantID, cfg.tenantsInContext)
			}

			// Act
			result := HasPermissionsFor(ctx, cfg.tenantsInput...)

			// Assert
			assert.Equal(t, cfg.expectedRes, result)
		})
	}
}
