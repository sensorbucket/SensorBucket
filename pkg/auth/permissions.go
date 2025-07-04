package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var ErrPermissionInvalid = web.NewError(http.StatusBadRequest, "permission value is invalid", "ERR_PERMISSION_INVALID")

type (
	// Singular permission, already validated
	Permission  string
	Permissions []Permission
)

// Permissions
const (
	// Device permissions
	READ_DEVICES  Permission = "READ_DEVICES"
	WRITE_DEVICES Permission = "WRITE_DEVICES"

	// API Key permissions
	READ_API_KEYS  Permission = "READ_API_KEYS"
	WRITE_API_KEYS Permission = "WRITE_API_KEYS"

	// Tenant permissions
	READ_TENANTS  Permission = "READ_TENANTS"
	WRITE_TENANTS Permission = "WRITE_TENANTS"

	// Pipelines
	READ_PIPELINES  Permission = "READ_PIPELINES"
	WRITE_PIPELINES Permission = "WRITE_PIPELINES"

	// Measurement permissions
	READ_MEASUREMENTS  Permission = "READ_MEASUREMENTS"
	WRITE_MEASUREMENTS Permission = "WRITE_MEASUREMENTS"

	// Tracing permissions
	READ_TRACING Permission = "READ_TRACING"

	// User worker permissions
	READ_USER_WORKERS  Permission = "READ_USER_WORKERS"
	WRITE_USER_WORKERS Permission = "WRITE_USER_WORKERS"

	READ_PROJECTS  Permission = "READ_PROJECTS"
	WRITE_PROJECTS Permission = "WRITE_PROJECTS"
)

var allPermissions = Permissions{
	READ_DEVICES,
	WRITE_DEVICES,
	READ_API_KEYS,
	WRITE_API_KEYS,
	READ_TENANTS,
	WRITE_TENANTS,
	READ_PIPELINES,
	WRITE_PIPELINES,
	READ_MEASUREMENTS,
	WRITE_MEASUREMENTS,
	READ_TRACING,
	READ_USER_WORKERS,
	WRITE_USER_WORKERS,
	READ_PROJECTS,
	WRITE_PROJECTS,
}

func AllPermissions() Permissions {
	return allPermissions
}

var stringPermissionMap = lo.SliceToMap(allPermissions, func(item Permission) (string, Permission) {
	return string(item), item
})

func (this Permissions) Fulfills(that Permissions) error {
	_, missing := lo.Difference(this, that)
	if len(missing) > 0 {
		return ErrUnauthorized
	}
	return nil
}

func SringToPermission(str string) (Permission, bool) {
	p, ok := stringPermissionMap[str]
	if !ok {
		log.Printf("Tried converting non-existant string to permission: %s\n", str)
	}
	return p, ok
}

func StringsToPermissions(keys []string) (Permissions, error) {
	permissions := make([]Permission, 0, len(keys))
	for _, str := range keys {
		permission, ok := SringToPermission(str)
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrPermissionInvalid, str)
		}
		permissions = append(permissions, permission)
	}
	return permissions, nil
}

func (permission Permission) String() string {
	return string(permission)
}

func (permissions *Permissions) UnmarshalJSON(data []byte) error {
	strings := []string{}
	if err := json.Unmarshal(data, &strings); err != nil {
		return fmt.Errorf("could not unmarshal permissions: %w", err)
	}
	perms, err := StringsToPermissions(strings)
	if err != nil {
		return fmt.Errorf("could not unmarshal permissions: %w", err)
	}
	*permissions = perms
	return nil
}

func (permissions Permissions) Validate() error {
	_, invalidPermissions := lo.Difference(allPermissions, permissions)
	if len(invalidPermissions) > 0 {
		return fmt.Errorf("%w: %s", ErrPermissionInvalid, strings.Join(lo.Map(invalidPermissions, func(item Permission, index int) string { return string(item) }), ", "))
	}
	return nil
}
