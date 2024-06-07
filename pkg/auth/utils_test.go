package auth

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustBeTenant(t *testing.T) {
	testCases := []struct {
		desc             string
		contextTenantID  *int64
		requiredTenantID int64
		expectedError    error
	}{
		{
			desc:             "No tenant in context should error",
			contextTenantID:  nil,
			requiredTenantID: 10,
			expectedError:    ErrUnauthorized,
		},
		{
			desc:             "Correct tenant in context",
			contextTenantID:  lo.ToPtr[int64](10),
			requiredTenantID: 10,
			expectedError:    nil,
		},
		{
			desc:             "Incorrect tenant in context",
			contextTenantID:  lo.ToPtr[int64](13),
			requiredTenantID: 10,
			expectedError:    ErrForbidden,
		},
		{
			desc:             "Required is tenant 0, context is nil",
			contextTenantID:  nil,
			requiredTenantID: 0,
			expectedError:    ErrUnauthorized,
		},
		{
			desc:             "Required is tenant 0, context is set",
			contextTenantID:  lo.ToPtr[int64](13),
			requiredTenantID: 0,
			expectedError:    ErrForbidden,
		},
		{
			desc:             "Context is 0, required is set",
			contextTenantID:  lo.ToPtr[int64](0),
			requiredTenantID: 5,
			expectedError:    ErrUnauthorized,
		},
		{
			desc:             "Context is 0, required is 0",
			contextTenantID:  lo.ToPtr[int64](0),
			requiredTenantID: 0,
			expectedError:    ErrUnauthorized,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.Background()
			if tC.contextTenantID != nil {
				ctx = setTenantID(ctx, *tC.contextTenantID)
			}

			err := MustBeTenant(ctx, tC.requiredTenantID)
			if tC.expectedError != nil {
				require.Error(t, err)
			}
			assert.ErrorIs(t, err, tC.expectedError)
		})
	}
}

func TestMustHaveTenantPermissions(t *testing.T) {
	testCases := []struct {
		desc                string
		contextTenantID     *int64
		contextPermissions  Permissions
		requiredTenantID    int64
		requiredPermissions Permissions
		expectedError       error
	}{
		{
			desc:                "No tenant and no perms should error",
			contextTenantID:     nil,
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       ErrUnauthorized,
		},
		{
			desc:                "Wrong tenant without perms should error",
			contextTenantID:     lo.ToPtr[int64](10),
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       ErrForbidden,
		},
		{
			desc:                "Wrong tenant with correct perms should error",
			contextTenantID:     lo.ToPtr[int64](10),
			contextPermissions:  Permissions{READ_DEVICES, WRITE_DEVICES},
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       ErrForbidden,
		},
		{
			desc:                "Correct tenant without perms should error",
			contextTenantID:     lo.ToPtr[int64](15),
			contextPermissions:  Permissions{},
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       ErrForbidden,
		},
		{
			desc:                "Correct tenant with partial correct perms should error",
			contextTenantID:     lo.ToPtr[int64](15),
			contextPermissions:  Permissions{READ_DEVICES},
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       ErrForbidden,
		},
		{
			desc:                "Correct tenant with correct perms should not error",
			contextTenantID:     lo.ToPtr[int64](15),
			contextPermissions:  Permissions{READ_DEVICES, WRITE_DEVICES},
			requiredTenantID:    15,
			requiredPermissions: Permissions{READ_DEVICES, WRITE_DEVICES},
			expectedError:       nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.Background()
			if tC.contextTenantID != nil {
				ctx = setTenantID(ctx, *tC.contextTenantID)
			}
			ctx = setPermissions(ctx, tC.contextPermissions)

			err := MustHaveTenantPermissions(ctx, tC.requiredTenantID, tC.requiredPermissions)
			if tC.expectedError != nil {
				require.Error(t, err)
			}
			assert.ErrorIs(t, err, tC.expectedError)
		})
	}
}

func TestGetTenant(t *testing.T) {
	// Arrange
	type testCase struct {
		tenantInContext *int64
		expectedRes     int64
		expectedErr     error
	}

	scenarios := map[string]testCase{
		"no tenants in context": {
			tenantInContext: nil,
			expectedRes:     0,
			expectedErr:     ErrNoTenantIDFound,
		},
		"with tenant in context": {
			tenantInContext: lo.ToPtr[int64](143),
			expectedRes:     143,
			expectedErr:     nil,
		},
		"with 0 in context should return error": {
			tenantInContext: lo.ToPtr[int64](0),
			expectedRes:     0,
			expectedErr:     ErrNoTenantIDFound,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			ctx := context.Background()
			if cfg.tenantInContext != nil {
				ctx = setTenantID(ctx, *cfg.tenantInContext)
			}

			// Act
			result, err := GetTenant(ctx)

			// Assert
			assert.Equal(t, cfg.expectedRes, result)
			assert.ErrorIs(t, err, cfg.expectedErr)
		})
	}
}
