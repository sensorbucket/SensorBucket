package auth

import (
	"context"
)

// Checks if the given context contains said permissions
// returns nil if all is OK
func MustHavePermissions(c context.Context, permissions ...permission) error {
	if len(permissions) == 0 {
		return ErrNoPermissionsToCheck
	}
	permissionsFromContext, ok := fromRequestContext[[]permission](c, PermissionsKey)
	if !ok {
		return ErrNoPermissions
	}
	for _, p := range permissions {
		found := false
		for _, fromContext := range permissionsFromContext {
			if p == fromContext {
				found = true
			}
		}
		if !found {
			return ErrPermissionsNotGranted
		}
	}
	return nil
}

func HasRole(ctx context.Context, r role) bool {
	permissionsFromContext, ok := fromRequestContext[[]permission](ctx, PermissionsKey)
	if !ok {
		return false
	}
	return r.contains(permissionsFromContext)
}

func GetUser(ctx context.Context) (int64, error) {
	val, ok := fromRequestContext[int64](ctx, UserIdKey)
	if !ok {
		return -1, ErrNoUserId
	}
	return val, nil
}

func GetTenants(ctx context.Context) ([]int64, error) {
	val, ok := fromRequestContext[[]int64](ctx, CurrentTenantIdKey)
	if !ok {
		return nil, ErrNoTenantIdFound
	}
	return val, nil
}

func HasPermissionsFor(ctx context.Context, tenantIds ...int64) bool {
	tenants, err := GetTenants(ctx)
	if err != nil {
		return false
	}
	if len(tenantIds) == 0 {
		return false
	}
	for _, t := range tenants {
		found := false
		for _, reqTenant := range tenantIds {
			if t == reqTenant {
				// Found tenant in the grant
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func fromRequestContext[T any](c context.Context, key string) (T, bool) {
	val, ok := c.Value(key).(T)
	return val, ok
}
