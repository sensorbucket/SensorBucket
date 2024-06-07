package auth

import (
	"context"
	"errors"
	"fmt"
)

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxTenantID
	ctxPermissions
	ctxAccessToken
)

var (
	ErrInvalidContext = errors.New("invalid auth context")
	ErrContextMissing = errors.New("missing auth context")
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

func setAccessToken(ctx context.Context, accessToken string) context.Context {
	return context.WithValue(ctx, ctxAccessToken, accessToken)
}

func GetTenant(ctx context.Context) (int64, error) {
	value := ctx.Value(ctxTenantID)
	if value == nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidContext, ErrNoTenantIDFound)
	}

	typedValue, ok := value.(int64)
	if !ok {
		return 0, fmt.Errorf("%w: TenantID value is wrong type %T", ErrInvalidContext, value)
	}

	if typedValue == 0 {
		return 0, fmt.Errorf("%w: %w", ErrInvalidContext, ErrNoTenantIDFound)
	}

	return typedValue, nil
}

func GetUser(ctx context.Context) (string, error) {
	value := ctx.Value(ctxUserID)
	if value == nil {
		return "", fmt.Errorf("%w: %w", ErrContextMissing, ErrNoUserID)
	}

	typedValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: UserID value is wrong type %T", ErrInvalidContext, value)
	}

	if typedValue == "" {
		return "", fmt.Errorf("%w: %w", ErrContextMissing, ErrNoUserID)
	}

	return typedValue, nil
}

func GetPermissions(ctx context.Context) (Permissions, error) {
	value := ctx.Value(ctxPermissions)
	if value == nil {
		return Permissions{}, fmt.Errorf("%w: %w", ErrInvalidContext, ErrNoPermissions)
	}

	typedValue, ok := value.(Permissions)
	if !ok {
		return Permissions{}, fmt.Errorf("%w: TenantID value is wrong type %T", ErrInvalidContext, value)
	}

	return typedValue, nil
}

func GetAccessToken(ctx context.Context) (string, error) {
	value := ctx.Value(ctxAccessToken)
	if value == nil {
		return "", fmt.Errorf("%w: %w", ErrContextMissing, ErrNoAccessToken)
	}

	typedValue, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("%w: AccessToken value is wrong type %T", ErrInvalidContext, value)
	}

	if typedValue == "" {
		return "", fmt.Errorf("%w: %w", ErrContextMissing, ErrNoAccessToken)
	}

	return typedValue, nil
}
