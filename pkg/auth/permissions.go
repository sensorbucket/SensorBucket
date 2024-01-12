package auth

import "fmt"

var (
	// Device permissions
	READ_DEVICES  permission = "READ_DEVICES"
	WRITE_DEVICES permission = "WRITE_DEVICES"

	// API Key permissions
	READ_API_KEYS  permission = "READ_API_KEYS"
	WRITE_API_KEYS permission = "WRITE_API_KEYS"

	// Tenant permissions
	READ_TENANTS  permission = "READ_TENANTS"
	WRITE_TENANTS permission = "WRITE_TENANTS"

	// Measurement permissions
	READ_MEASUREMENTS  permission = "READ_MEASUREMENTS"
	WRITE_MEASUREMENTS permission = "WRITE_MEASUREMENTS"

	// Tracing permissions
	READ_TRACING permission = "READ_TRACING"

	// User worker permissions
	READ_USER_WORKERS  permission = "READ_USER_WORKERS"
	WRITE_USER_WORKERS permission = "WRITE_USER_WORKERS"
)

var allowedPermissions = []permission{
	READ_DEVICES,
	WRITE_DEVICES,
	READ_API_KEYS,
	WRITE_API_KEYS,
	READ_TENANTS,
	WRITE_TENANTS,
	READ_MEASUREMENTS,
	WRITE_MEASUREMENTS,
	READ_TRACING,
	READ_USER_WORKERS,
	WRITE_USER_WORKERS,
}

func PermissionsValid(permissionStrings []string) bool {
	for _, perm := range permissionStrings {
		if permission(perm).Valid() != nil {
			return false
		}
	}
	return len(permissionStrings) > 0
}

func NewRole(permissions ...permission) role {
	return role(permissions)
}

type permission string

func (p permission) String() string {
	return string(p)
}

func (p permission) Valid() error {
	for _, allowed := range allowedPermissions {
		if allowed == p {
			return nil
		}
	}
	return fmt.Errorf("%s is not a valid permission", p)
}

type role []permission

func (r role) contains(permissions []permission) bool {
	for _, rolePermission := range r {
		found := false
		for _, p := range permissions {
			if rolePermission == p {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}
