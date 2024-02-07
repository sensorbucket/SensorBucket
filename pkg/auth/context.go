package auth

import "context"

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxTenantID
	ctxPermissions
)

func setUserID(ctx context.Context, userID int64) context.Context {
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

func GetUser(ctx context.Context) (int64, error) {
	val, ok := fromContext[int64](ctx, ctxUserID)
	if !ok {
		return 0, ErrNoUserID
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
	val, ok := ctx.Value(key).(T)
	return val, ok
}

type contextBuilder struct {
	c context.Context
}

func (cb *contextBuilder) With(key ctxKey, value any) *contextBuilder {
	cb.c = context.WithValue(cb.c, key, value)
	return cb
}

func (cb *contextBuilder) Finish() context.Context {
	return cb.c
}
