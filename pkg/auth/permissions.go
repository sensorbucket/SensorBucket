package auth

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

	// Roles
	DEVICE_MANAGER role = role([]permission{
		READ_DEVICES,
		WRITE_DEVICES,
	})
	API_KEY_MANAGER role = role([]permission{
		READ_API_KEYS,
		WRITE_API_KEYS,
	})
)

type permission string

func (p permission) String() string {
	return string(p)
}

type role []permission

func (r role) Permissions() []permission {
	return []permission(r)
}

func (r role) Contains(permissions []permission) bool {
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
