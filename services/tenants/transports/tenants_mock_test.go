// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package tenantstransports

import (
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
	"sync"
)

// Ensure, that tenantServiceMock does implement tenantService.
// If this is not the case, regenerate this file with moq.
var _ tenantService = &tenantServiceMock{}

// tenantServiceMock is a mock implementation of tenantService.
//
//	func TestSomethingThatUsestenantService(t *testing.T) {
//
//		// make and configure a mocked tenantService
//		mockedtenantService := &tenantServiceMock{
//			ArchiveTenantFunc: func(tenantID int64) error {
//				panic("mock out the ArchiveTenant method")
//			},
//			CreateNewTenantFunc: func(tenant tenants.TenantDTO) (*tenants.TenantDTO, error) {
//				panic("mock out the CreateNewTenant method")
//			},
//			ListTenantsFunc: func(filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.TenantDTO], error) {
//				panic("mock out the ListTenants method")
//			},
//		}
//
//		// use mockedtenantService in code that requires tenantService
//		// and then make assertions.
//
//	}
type tenantServiceMock struct {
	// ArchiveTenantFunc mocks the ArchiveTenant method.
	ArchiveTenantFunc func(tenantID int64) error

	// CreateNewTenantFunc mocks the CreateNewTenant method.
	CreateNewTenantFunc func(tenant tenants.TenantDTO) (*tenants.TenantDTO, error)

	// ListTenantsFunc mocks the ListTenants method.
	ListTenantsFunc func(filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.TenantDTO], error)

	// calls tracks calls to the methods.
	calls struct {
		// ArchiveTenant holds details about calls to the ArchiveTenant method.
		ArchiveTenant []struct {
			// TenantID is the tenantID argument value.
			TenantID int64
		}
		// CreateNewTenant holds details about calls to the CreateNewTenant method.
		CreateNewTenant []struct {
			// Tenant is the tenant argument value.
			Tenant tenants.TenantDTO
		}
		// ListTenants holds details about calls to the ListTenants method.
		ListTenants []struct {
			// Filter is the filter argument value.
			Filter tenants.Filter
			// P is the p argument value.
			P pagination.Request
		}
	}
	lockArchiveTenant   sync.RWMutex
	lockCreateNewTenant sync.RWMutex
	lockListTenants     sync.RWMutex
}

// ArchiveTenant calls ArchiveTenantFunc.
func (mock *tenantServiceMock) ArchiveTenant(tenantID int64) error {
	if mock.ArchiveTenantFunc == nil {
		panic("tenantServiceMock.ArchiveTenantFunc: method is nil but tenantService.ArchiveTenant was just called")
	}
	callInfo := struct {
		TenantID int64
	}{
		TenantID: tenantID,
	}
	mock.lockArchiveTenant.Lock()
	mock.calls.ArchiveTenant = append(mock.calls.ArchiveTenant, callInfo)
	mock.lockArchiveTenant.Unlock()
	return mock.ArchiveTenantFunc(tenantID)
}

// ArchiveTenantCalls gets all the calls that were made to ArchiveTenant.
// Check the length with:
//
//	len(mockedtenantService.ArchiveTenantCalls())
func (mock *tenantServiceMock) ArchiveTenantCalls() []struct {
	TenantID int64
} {
	var calls []struct {
		TenantID int64
	}
	mock.lockArchiveTenant.RLock()
	calls = mock.calls.ArchiveTenant
	mock.lockArchiveTenant.RUnlock()
	return calls
}

// CreateNewTenant calls CreateNewTenantFunc.
func (mock *tenantServiceMock) CreateNewTenant(tenant tenants.TenantDTO) (*tenants.TenantDTO, error) {
	if mock.CreateNewTenantFunc == nil {
		panic("tenantServiceMock.CreateNewTenantFunc: method is nil but tenantService.CreateNewTenant was just called")
	}
	callInfo := struct {
		Tenant tenants.TenantDTO
	}{
		Tenant: tenant,
	}
	mock.lockCreateNewTenant.Lock()
	mock.calls.CreateNewTenant = append(mock.calls.CreateNewTenant, callInfo)
	mock.lockCreateNewTenant.Unlock()
	return mock.CreateNewTenantFunc(tenant)
}

// CreateNewTenantCalls gets all the calls that were made to CreateNewTenant.
// Check the length with:
//
//	len(mockedtenantService.CreateNewTenantCalls())
func (mock *tenantServiceMock) CreateNewTenantCalls() []struct {
	Tenant tenants.TenantDTO
} {
	var calls []struct {
		Tenant tenants.TenantDTO
	}
	mock.lockCreateNewTenant.RLock()
	calls = mock.calls.CreateNewTenant
	mock.lockCreateNewTenant.RUnlock()
	return calls
}

// ListTenants calls ListTenantsFunc.
func (mock *tenantServiceMock) ListTenants(filter tenants.Filter, p pagination.Request) (*pagination.Page[tenants.TenantDTO], error) {
	if mock.ListTenantsFunc == nil {
		panic("tenantServiceMock.ListTenantsFunc: method is nil but tenantService.ListTenants was just called")
	}
	callInfo := struct {
		Filter tenants.Filter
		P      pagination.Request
	}{
		Filter: filter,
		P:      p,
	}
	mock.lockListTenants.Lock()
	mock.calls.ListTenants = append(mock.calls.ListTenants, callInfo)
	mock.lockListTenants.Unlock()
	return mock.ListTenantsFunc(filter, p)
}

// ListTenantsCalls gets all the calls that were made to ListTenants.
// Check the length with:
//
//	len(mockedtenantService.ListTenantsCalls())
func (mock *tenantServiceMock) ListTenantsCalls() []struct {
	Filter tenants.Filter
	P      pagination.Request
} {
	var calls []struct {
		Filter tenants.Filter
		P      pagination.Request
	}
	mock.lockListTenants.RLock()
	calls = mock.calls.ListTenants
	mock.lockListTenants.RUnlock()
	return calls
}