package tenants_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestCreateParentTenantDoesNotExist(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(132), id)
			return nil, tenants.ErrTenantNotFound
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	parent := int64(132)
	_, err := s.CreateNewTenant(context.Background(), tenants.CreateTenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, tenants.ErrTenantNotFound)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
}

func TestCreateParentTenantCantBeRetrieved(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("some weird database error has occurred")
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(675), id)
			return nil, expErr
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(context.Background(), tenants.CreateTenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
}

func TestCreateParentTenantIsNotActive(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &tenants.Tenant{
				ID:    675,
				State: tenants.Archived,
			}, nil
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(context.Background(), tenants.CreateTenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, tenants.ErrTenantNotActive)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
}

func TestCreateErrorOccurs(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error!")
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &tenants.Tenant{
				ID:    675,
				State: tenants.Active,
			}, nil
		},
		CreateFunc: func(tenant *tenants.Tenant) error {
			return expErr
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	parent := int64(675)
	_, err := s.CreateNewTenant(context.Background(), tenants.CreateTenantDTO{
		ParentID: &parent,
	})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
	assert.Len(t, store.CreateCalls(), 1)
}

func TestCreateCreatesNewTenant(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(675), id)
			return &tenants.Tenant{
				ID:    675,
				State: tenants.Active,
			}, nil
		},
		CreateFunc: func(tenant *tenants.Tenant) error {
			return nil
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	parent := int64(675)
	dto, err := s.CreateNewTenant(context.Background(), tenants.CreateTenantDTO{
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
	assert.Equal(t, tenants.CreateTenantDTO{
		Name:    "blabla",
		Address: "somewhere nice",
		ZipCode: "no clue",
		City:    "some place",
		// ChamberOfCommerceID: "ideee",
		// HeadquarterID:       "hqid",
		ParentID: &parent,
	}, dto)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
	assert.Len(t, store.CreateCalls(), 1)
}

func TestArchiveTenantErrorOccursWhileRetrievingTenant(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error")
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return nil, expErr
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	err := s.ArchiveTenant(context.Background(), 43124)

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
}

func TestArchiveTenantTenantIsAlreadyArchived(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &tenants.Tenant{
				State: tenants.Archived,
			}, nil
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	err := s.ArchiveTenant(context.Background(), 43124)

	// Assert
	assert.ErrorIs(t, err, tenants.ErrTenantNotActive)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
}

func TestArchiveTenantUpdateErrors(t *testing.T) {
	// Arrange
	expErr := fmt.Errorf("weird error")
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &tenants.Tenant{
				State: tenants.Active,
			}, nil
		},
		UpdateFunc: func(tenant *tenants.Tenant) error {
			assert.Equal(t, tenants.Archived, tenant.State)
			return expErr
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	err := s.ArchiveTenant(context.Background(), 43124)

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
	assert.Len(t, store.UpdateCalls(), 1)
}

func TestArchiveTenantUpdatesTenantWithArchivedState(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(43124), id)
			return &tenants.Tenant{
				State: tenants.Active,
			}, nil
		},
		UpdateFunc: func(tenant *tenants.Tenant) error {
			assert.Equal(t, tenants.Archived, tenant.State)
			return nil
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	err := s.ArchiveTenant(context.Background(), 43124)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, store.GetTenantByIDCalls(), 1)
	assert.Len(t, store.UpdateCalls(), 1)
}

func TestListTenantsReturnsList(t *testing.T) {
	// Arrange
	store := &TenantStoreMock{
		ListFunc: func(filter tenants.StoreFilter, request pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error) {
			return &pagination.Page[tenants.CreateTenantDTO]{
				Cursor: "blabla",
				Data: []tenants.CreateTenantDTO{
					{
						Name: "123adsz",
					},
				},
			}, nil
		},
	}
	s := tenants.NewTenantService(store, nil)

	// Act
	res, err := s.ListTenants(context.Background(), tenants.Filter{}, pagination.Request{})

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
		ListFunc: func(filter tenants.StoreFilter, request pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error) {
			return nil, expErr
		},
	}
	s := tenants.NewTenantService(&store, nil)

	// Act
	res, err := s.ListTenants(context.Background(), tenants.Filter{}, pagination.Request{})

	// Assert
	assert.ErrorIs(t, err, expErr)
	assert.Nil(t, res)
	assert.Len(t, store.ListCalls(), 1)
}

func TestCreateTenantMember(t *testing.T) {
	tenant := tenants.Tenant{
		ID:                  15,
		Name:                "",
		Address:             "",
		ZipCode:             "",
		City:                "",
		ChamberOfCommerceID: nil,
		HeadquarterID:       nil,
		ArchiveTime:         nil,
		State:               tenants.Active,
		Logo:                nil,
		ParentID:            nil,
	}
	userID := "123123"
	permissions := auth.Permissions{auth.WRITE_DEVICES, auth.READ_DEVICES}
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			return &tenant, nil
		},
		SaveMemberFunc: func(tenantID int64, member *tenants.Member) error {
			return nil
		},
		GetTenantMemberFunc: func(tenantID int64, userID string) (*tenants.Member, error) {
			return nil, tenants.ErrTenantMemberNotFound
		},
	}
	userValidator := &UserValidatorMock{
		UserByIDExistsFunc: func(ctx context.Context, tenantID int64, userID string) error {
			return nil
		},
	}
	service := tenants.NewTenantService(store, userValidator)

	err := service.AddTenantMember(context.Background(), tenant.ID, userID, permissions)
	require.NoError(t, err)

	require.Len(t, store.calls.GetTenantByID, 1)
	require.Len(t, store.calls.SaveMember, 1)
	assert.Equal(t, store.calls.GetTenantByID[0].ID, tenant.ID)
	member := store.calls.SaveMember[0].Member
	assert.Equal(t, tenant.ID, store.calls.SaveMember[0].TenantID)
	assert.Equal(t, userID, member.UserID)
	assert.Equal(t, permissions, member.Permissions)
}

func TestTenantAddMemberShouldErrorWithInvalidPermissions(t *testing.T) {
	tenant := tenants.Tenant{
		ID:                  15,
		Name:                "",
		Address:             "",
		ZipCode:             "",
		City:                "",
		ChamberOfCommerceID: nil,
		HeadquarterID:       nil,
		ArchiveTime:         nil,
		State:               tenants.Active,
		Logo:                nil,
		ParentID:            nil,
	}
	userID := "123123"
	permissions := auth.Permissions{auth.WRITE_DEVICES, auth.Permission("1283719823")}
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			return &tenant, nil
		},
		SaveMemberFunc: func(tenantID int64, member *tenants.Member) error {
			return nil
		},
		GetTenantMemberFunc: func(tenantID int64, userID string) (*tenants.Member, error) {
			return nil, tenants.ErrTenantMemberNotFound
		},
	}
	userValidator := &UserValidatorMock{
		UserByIDExistsFunc: func(ctx context.Context, tenantID int64, userID string) error {
			return nil
		},
	}
	service := tenants.NewTenantService(store, userValidator)

	err := service.AddTenantMember(context.Background(), tenant.ID, userID, permissions)
	assert.ErrorIs(t, err, auth.ErrPermissionInvalid)
	assert.Len(t, store.calls.GetTenantByID, 0)
	assert.Len(t, store.calls.SaveMember, 0)
}

func TestTenantModifyMemberShouldErrorWithInvalidPermissions(t *testing.T) {
	tenant := tenants.Tenant{
		ID:                  15,
		Name:                "",
		Address:             "",
		ZipCode:             "",
		City:                "",
		ChamberOfCommerceID: nil,
		HeadquarterID:       nil,
		ArchiveTime:         nil,
		State:               tenants.Active,
		Logo:                nil,
		ParentID:            nil,
	}
	userID := "123123"
	origPermissions := auth.Permissions{auth.WRITE_DEVICES}
	newPermissions := auth.Permissions{auth.WRITE_DEVICES, auth.Permission("1283719823")}
	member := tenants.Member{
		UserID:      userID,
		Permissions: origPermissions,
	}
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			return &tenant, nil
		},
		GetTenantMemberFunc: func(tenantID int64, userID string) (*tenants.Member, error) {
			return &member, nil
		},
		SaveMemberFunc: func(tenantID int64, member *tenants.Member) error {
			return nil
		},
	}
	userValidator := &UserValidatorMock{
		UserByIDExistsFunc: func(ctx context.Context, tenantID int64, userID string) error {
			return nil
		},
	}
	service := tenants.NewTenantService(store, userValidator)

	err := service.ModifyMemberPermissions(context.Background(), tenant.ID, userID, newPermissions)
	assert.ErrorIs(t, err, auth.ErrPermissionInvalid)
	assert.Len(t, store.calls.GetTenantByID, 0)
	assert.Len(t, store.calls.SaveMember, 0)
}

func TestTenantAddMemberShouldErrorIfUserDoesNotExist(t *testing.T) {
	tenant := tenants.Tenant{
		ID:                  15,
		Name:                "",
		Address:             "",
		ZipCode:             "",
		City:                "",
		ChamberOfCommerceID: nil,
		HeadquarterID:       nil,
		ArchiveTime:         nil,
		State:               tenants.Active,
		Logo:                nil,
		ParentID:            nil,
	}
	userID := "123123"
	permissions := auth.Permissions{auth.WRITE_DEVICES}
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			return &tenant, nil
		},
		SaveMemberFunc: func(tenantID int64, member *tenants.Member) error {
			return nil
		},
		GetTenantMemberFunc: func(tenantID int64, userID string) (*tenants.Member, error) {
			return nil, tenants.ErrTenantMemberNotFound
		},
	}
	userValidator := &UserValidatorMock{
		UserByIDExistsFunc: func(ctx context.Context, tenantID int64, userID string) error {
			return errors.New("TODO: UserNotFoundError")
		},
	}
	service := tenants.NewTenantService(store, userValidator)

	err := service.AddTenantMember(context.Background(), tenant.ID, userID, permissions)

	assert.Error(t, err)
	assert.Len(t, store.calls.GetTenantByID, 1)
	assert.Len(t, userValidator.calls.UserByIDExists, 1)
	assert.Len(t, store.calls.SaveMember, 0)
}

func TestTenantModifyMemberChangesPermissions(t *testing.T) {
	tenant := tenants.Tenant{
		ID:                  15,
		Name:                "",
		Address:             "",
		ZipCode:             "",
		City:                "",
		ChamberOfCommerceID: nil,
		HeadquarterID:       nil,
		ArchiveTime:         nil,
		State:               tenants.Active,
		Logo:                nil,
		ParentID:            nil,
	}
	userID := "123123"
	origPermissions := auth.Permissions{auth.WRITE_DEVICES}
	newPermissions := auth.Permissions{auth.WRITE_DEVICES, auth.READ_API_KEYS}
	member := tenants.Member{
		UserID:      userID,
		Permissions: origPermissions,
	}
	store := &TenantStoreMock{
		GetTenantByIDFunc: func(id int64) (*tenants.Tenant, error) {
			return &tenant, nil
		},
		GetTenantMemberFunc: func(tenantID int64, userID string) (*tenants.Member, error) {
			return &member, nil
		},
		SaveMemberFunc: func(tenantID int64, member *tenants.Member) error {
			return nil
		},
	}
	service := tenants.NewTenantService(store, nil)

	err := service.ModifyMemberPermissions(context.Background(), tenant.ID, userID, newPermissions)
	assert.NoError(t, err)
	require.Len(t, store.calls.GetTenantByID, 1)
	require.Len(t, store.calls.SaveMember, 1)
	assert.Equal(t, store.calls.GetTenantByID[0].ID, tenant.ID)
	calledMember := store.calls.SaveMember[0].Member
	assert.Equal(t, tenant.ID, store.calls.SaveMember[0].TenantID)
	assert.Equal(t, userID, calledMember.UserID)
	assert.Equal(t, newPermissions, calledMember.Permissions)
}
