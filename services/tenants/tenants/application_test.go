package tenants_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	. "sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestCreateParentTenantDoesNotExist(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(132), id)
			return nil, ErrTenantNotFound
		},
	}
	s := NewTenantService(store)

	// Act
	parent := int64(132)
	_, err := s.CreateNewTenant(TenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, ErrTenantNotFound)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
}

func TestCreateParentTenantCantBeRetrieved(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("some weird database error has occurred")
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(675), id)
			return nil, expErr
		},
	}
	s := NewTenantService(store)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(TenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
}

func TestCreateParentTenantIsNotActive(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &Tenant{
				ID:    675,
				State: Archived,
			}, nil
		},
	}
	s := NewTenantService(store)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(TenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, ErrTenantNotActive)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
}

func TestCreateErrorOccurs(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error!")
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &Tenant{
				ID:    675,
				State: Active,
			}, nil
		},
		CreateFunc: func(tenant *Tenant) error {
			return expErr
		},
	}
	s := NewTenantService(store)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(TenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
	assert.Len(t, store.CreateCalls(), 1)
}

func TestCreateCreatesNewTenant(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &Tenant{
				ID:    675,
				State: Active,
			}, nil
		},
		CreateFunc: func(tenant *Tenant) error {
			return nil
		},
	}
	s := NewTenantService(store)

	// Act
	parent := int64(675)
	dto, err := s.CreateNewTenant(TenantDTO{
		Name:    "blabla",
		Address: "somewhere nice",
		ZipCode: "no clue",
		City:    "some place",
		// ChamberOfCommerceID: "ideee",
		// HeadquarterID:       "hqid",
		ParentID: &parent,
	})

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, TenantDTO{
		Name:    "blabla",
		Address: "somewhere nice",
		ZipCode: "no clue",
		City:    "some place",
		// ChamberOfCommerceID: "ideee",
		// HeadquarterID:       "hqid",
		ParentID: &parent,
	}, dto)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
	assert.Len(t, store.CreateCalls(), 1)
}

func TestArchiveTenantErrorOccursWhileRetrievingTenant(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error")
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return nil, expErr
		},
	}
	s := NewTenantService(store)

	// Act
	err := s.ArchiveTenant(43124)

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
}

func TestArchiveTenantTenantIsAlreadyArchived(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &Tenant{
				State: Archived,
			}, nil
		},
	}
	s := NewTenantService(store)

	// Act
	err := s.ArchiveTenant(43124)

	// Assert
	assert.ErrorIs(t, err, ErrTenantNotActive)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
}

func TestArchiveTenantUpdateErrors(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error")
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &Tenant{
				State: Active,
			}, nil
		},
		UpdateFunc: func(tenant *Tenant) error {
			assert.Equal(t, Archived, tenant.State)
			return expErr
		},
	}
	s := NewTenantService(store)

	// Act
	err := s.ArchiveTenant(43124)

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
	assert.Len(t, store.UpdateCalls(), 1)
}

func TestArchiveTenantUpdatesTenantWithArchivedState(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIdFunc: func(id int64) (*Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &Tenant{
				State: Active,
			}, nil
		},
		UpdateFunc: func(tenant *Tenant) error {
			assert.Equal(t, Archived, tenant.State)
			return nil
		},
	}
	s := NewTenantService(store)

	// Act
	err := s.ArchiveTenant(43124)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, store.GetTenantByIdCalls(), 1)
	assert.Len(t, store.UpdateCalls(), 1)
}

func TestListTenantsReturnsList(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		ListFunc: func(filter Filter, request pagination.Request) (*pagination.Page[TenantDTO], error) {
			return &pagination.Page[TenantDTO]{
				Cursor: "blabla",
				Data: []TenantDTO{
					{
						Name: "123adsz",
					},
				},
			}, nil
		},
	}
	s := NewTenantService(store)

	// Act
	res, err := s.ListTenants(Filter{}, pagination.Request{})

	// Assert
	assert.Equal(t, "blabla", res.Cursor)
	assert.Len(t, res.Data, 1)
	assert.Equal(t, "123adsz", res.Data[0].Name)
	assert.NoError(t, err)
	assert.Len(t, store.ListCalls(), 1)
}

func TestListTenantsErrorOccursWhileRetrievingList(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error")
	store := TenantStoreMock{
		ListFunc: func(filter Filter, request pagination.Request) (*pagination.Page[TenantDTO], error) {
			return nil, expErr
		},
	}
	s := NewTenantService(&store)

	// Act
	res, err := s.ListTenants(Filter{}, pagination.Request{})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Nil(t, res)
	assert.Len(t, store.ListCalls(), 1)
}
