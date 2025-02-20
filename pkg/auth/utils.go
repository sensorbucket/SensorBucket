package auth

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Masterminds/squirrel"
)

func MustHavePermissions(ctx context.Context, required Permissions) error {
	permissions, err := GetPermissions(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if err := permissions.Fulfills(required); err != nil {
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	return nil
}

// MustHaveTenantPermissions is the only validating exported authentication and authorization method
// it requires the developer to supply both the tenant for whom this request must be and accompanying permissions
func MustHaveTenantPermissions(ctx context.Context, tenantID int64, required Permissions) error {
	if err := MustBeTenant(ctx, tenantID); err != nil {
		return err
	}
	permissions, err := GetPermissions(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if err := permissions.Fulfills(required); err != nil {
		return fmt.Errorf("%w: %w", ErrForbidden, err)
	}
	return nil
}

func MustBeTenant(ctx context.Context, tenantID int64) error {
	tenant, err := GetTenant(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrUnauthorized, err)
	}
	if tenant != tenantID {
		return ErrForbidden
	}
	return nil
}

var bearerLen = len("Bearer ") // Includes the space!
func StripBearer(str string) (string, bool) {
	if len(str) < bearerLen {
		return str, false
	}
	if !strings.EqualFold(str[:bearerLen], "bearer ") {
		return str, false
	}
	return str[bearerLen:], true
}

func CreateAuthenticatedContextForTESTING(ctx context.Context, sub string, tenantID int64, permissions Permissions) context.Context {
	ctx = setUserID(ctx, sub)
	ctx = setTenantID(ctx, tenantID)
	ctx = setPermissions(ctx, permissions)
	return ctx
}

type queryBuilders interface {
	squirrel.SelectBuilder | squirrel.DeleteBuilder | squirrel.UpdateBuilder | squirrel.StatementBuilderType
}

type pqts[T queryBuilders] interface {
	Where(pred any, args ...any) T
}

func ProtectedQuery[T queryBuilders](ctx context.Context, tenantIDColumn string, query pqts[T]) T {
	tenantID, err := GetTenant(ctx)
	if err != nil {
		log.Println("WARN: in pkg/auth/utils.go. Called ProtectedQuery without a tenant being set in the context")
		return query.Where("false")
	}
	return query.Where(squirrel.Eq{tenantIDColumn: tenantID})
}
