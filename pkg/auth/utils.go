package auth

import (
	"context"

	"github.com/samber/lo"
)

// Checks if the given context contains said permissions
// returns nil if all is OK
func MustHavePermissions(c context.Context, perm permission, permissions ...permission) error {
	permissions = append(permissions, perm)
	permissionsFromContext, ok := fromRequestContext[[]permission](c, ctxPermissions)
	if !ok {
		return ErrNoPermissions
	}
	if lo.Every(permissionsFromContext, permissions) {
		return nil
	}
	return ErrPermissionsNotGranted
}

func HasRole(ctx context.Context, r Role) bool {
	permissionsFromContext, ok := fromRequestContext[[]permission](ctx, ctxPermissions)
	if !ok {
		return false
	}
	if len(permissionsFromContext) == 0 {
		return false
	}
	if len(permissionsFromContext) > 1 {
		return r.HasPermissions(permissionsFromContext[0], permissionsFromContext...)
	}
	return r.HasPermissions(permissionsFromContext[0])
}

func GetUser(ctx context.Context) (int64, error) {
	val, ok := fromRequestContext[int64](ctx, ctxUserID)
	if !ok {
		return -1, ErrNoUserID
	}
	return val, nil
}

func GetTenants(ctx context.Context) ([]int64, error) {
	val, ok := fromRequestContext[[]int64](ctx, ctxCurrentTenantID)
	if !ok {
		return nil, ErrNoTenantIDFound
	}
	return val, nil
}

func HasPermissionsFor(ctx context.Context, tenantIDs ...int64) bool {
	tenants, err := GetTenants(ctx)
	if err != nil {
		return false
	}
	if len(tenantIDs) == 0 {
		return false
	}
	return lo.Every(tenants, tenantIDs)
}

func fromRequestContext[T any](c context.Context, key ctxKey) (T, bool) {
	val, ok := c.Value(key).(T)
	return val, ok
}
