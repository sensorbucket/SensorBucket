package auth

import (
	"context"
	"fmt"
)

// MustHaveTenantPermissions is the only validating exported authentication and authorization method
// it requires the developer to supply both the tenant for whom this request must be and accompanying permissions
func MustHaveTenantPermissions(ctx context.Context, tenantID int64, required Permissions) error {
	if err := mustBeTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	permissions, err := GetPermissions(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if err := permissions.Fulfills(required); err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	return nil
}

func mustBeTenant(ctx context.Context, tenantID int64) error {
	tenant, err := GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if tenant != tenantID {
		return ErrUnauthorized
	}
	return nil
}
