package auth

import "context"

// Checks if the given context contains said permissions, returns the tenant id for which the context is given the permissions
// returns nil if all is OK
func MustHavePermissions(c context.Context, permissions ...permission) (Grant, error) {
	if len(permissions) == 0 {
		return Grant{}, ErrNoPermissionsToCheck
	}
	tenantId, ok := fromRequestContext[int64](c, CurrentTenantIdKey)
	if !ok {
		return Grant{}, ErrNoTenantIdFound
	}
	permissionsFromContext, ok := fromRequestContext[[]permission](c, PermissionsKey)
	if !ok {
		return Grant{}, ErrNoPermissions
	}
	userId, ok := fromRequestContext[int64](c, UserIdKey)
	if !ok {
		return Grant{}, ErrNoUserId
	}
	for _, p := range permissions {
		found := false
		for _, fromContext := range permissionsFromContext {
			if p == fromContext {
				found = true
			}
		}
		if !found {
			return Grant{}, ErrPermissionsNotGranted
		}
	}
	return Grant{
		user: userId,
		tenants: []int64{
			tenantId,
		},
		permissions: permissionsFromContext,
	}, nil
}

type Grant struct {
	user        int64
	tenants     []int64
	permissions []permission
}

func (g *Grant) GetUser() int64 {
	return g.user
}

func (g *Grant) GetTenants() []int64 {
	return g.tenants
}

func (g *Grant) HasRole(r role) bool {
	return r.Contains(g.permissions)
}

func (g *Grant) HasPermissionsFor(tenantIds ...int64) bool {
	if len(tenantIds) == 0 {
		return false
	}
	for _, t := range g.tenants {
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
