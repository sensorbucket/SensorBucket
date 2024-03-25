package auth

import "context"

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxTenantID
	ctxPermissions
)

func setUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, ctxUserID, userID)
}

func setTenantID(ctx context.Context, tenantID int64) context.Context {
	return context.WithValue(ctx, ctxTenantID, tenantID)
}

func setPermissions(ctx context.Context, permissions Permissions) context.Context {
	return context.WithValue(ctx, ctxPermissions, permissions)
}

func GetTenant(ctx context.Context) (int64, error) {
	val, ok := fromContext[int64](ctx, ctxTenantID)
	if !ok || val == 0 {
		return 0, ErrNoTenantIDFound
	}
	return val, nil
}

func GetUser(ctx context.Context) (string, error) {
	val, ok := fromContext[string](ctx, ctxUserID)
	if !ok || val == "" {
		return "", ErrNoUserID
	}
	return val, nil
}

func GetPermissions(ctx context.Context) (Permissions, error) {
	val, ok := fromContext[Permissions](ctx, ctxPermissions)
	if !ok {
		return Permissions{}, ErrNoPermissions
	}
	return val, nil
}

func fromContext[T any](ctx context.Context, key ctxKey) (T, bool) {
	var val T
	var ok bool
	ival := ctx.Value(key)
	if ival == nil {
		return val, false
	}
	val, ok = ival.(T)
	return val, ok
}
