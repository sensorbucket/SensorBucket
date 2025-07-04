package sessions

import "sensorbucket.nl/sensorbucket/services/tenants/tenants"

type TenantStore = tenants.TenantStore

//go:generate moq -pkg sessions_test -out mock_test.go . UserPreferenceStore TenantStore
