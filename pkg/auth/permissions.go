package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/samber/lo"
)

var ErrPermissionInvalid = errors.New("permission value is invalid")

type PermissionSet interface {
	Permissions() []Permission
}

type Permission string

type Permissions []PermissionSet

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

	// Measurement permissions
	READ_MEASUREMENTS  Permission = "READ_MEASUREMENTS"
	WRITE_MEASUREMENTS Permission = "WRITE_MEASUREMENTS"

	// Tracing permissions
	READ_TRACING Permission = "READ_TRACING"

	// User worker permissions
	READ_USER_WORKERS  Permission = "READ_USER_WORKERS"
	WRITE_USER_WORKERS Permission = "WRITE_USER_WORKERS"
)

var allPermissions = Permissions{
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

func (p Permission) Permissions() []Permission {
	return []Permission{p}
}

func (gotten Permissions) Fulfills(required Permissions) error {
	flatGotten := gotten.Permissions()
	flatRequired := required.Permissions()
	missing, _ := lo.Difference(flatRequired, flatGotten)
	if len(missing) > 0 {
		return fmt.Errorf("missing: %v", missing)
	}
	return nil
}

func (p Permissions) Permissions() []Permission {
	return lo.Uniq(lo.Flatten(lo.Map(p, func(item PermissionSet, index int) []Permission { return item.Permissions() })))
}

func (p Permissions) Includes(other Permission) bool {
	return lo.IndexOf(p.Permissions(), other) != -1
}

func (p Permissions) Validate() error {
	invalidPermissions := lo.FilterMap(p.Permissions(), func(item Permission, _ int) (string, bool) {
		if err := item.Valid(); err != nil {
			return item.String(), true
		}
		return "", false
	})
	if len(invalidPermissions) > 0 {
		return fmt.Errorf("%w: %s", ErrPermissionInvalid, strings.Join(invalidPermissions, ", "))
	}
	return nil
}

func (s Permissions) String() string {
	return strings.Join(
		lo.Map(s.Permissions(), func(item Permission, _ int) string { return string(item) }),
		", ",
	)
}

func (s *Permissions) UnmarshalJSON(data []byte) error {
	var permissionSlice []Permission
	if err := json.Unmarshal(data, &permissionSlice); err != nil {
		return err
	}
	permissions := Permissions{}
	for _, p := range permissionSlice {
		permissions = append(permissions, p)
	}
	*s = permissions
	return nil
}

func (p Permission) String() string {
	return string(p)
}

func (p Permission) Valid() error {
	if allPermissions.Includes(p) {
		return nil
	}
	return fmt.Errorf("%w (value: %s)", ErrPermissionInvalid, p)
}

func AllPermissions() PermissionSet {
	return allPermissions
}
