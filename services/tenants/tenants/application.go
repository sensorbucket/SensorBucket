package tenants

import (
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type TenantService struct {
	tenantStore tenantStore
}

type Filter struct {
	Name  []string `json:"name"`
	State []State  `json:"state"`
}

type TenantDTO struct {
	ID                  int64          `json:"id"`
	Name                string         `json:"name"`
	Address             string         `json:"address"`
	ZipCode             string         `json:"zip_code"`
	City                string         `json:"city"`
	ChamberOfCommerceID *string        `json:"chamber_of_commerce_id"`
	HeadquarterID       *string        `json:"headquarter_id"`
	ArchiveTime         *time.Duration `json:"archive_time"`
	Logo                *string        `json:"logo"`
	ParentID            *int64         `json:"parent_tenant_id"`
	Permissions         []string       `json:"permissions"`
}

type AddMemberPermissionsDTO struct {
	UserID      int64    `json:"user_id"`
	TenantID    int64    `json:"tenant_id"`
	Permissions []string `json:"permissions"`
}

type MemberPermissionsAddedDTO struct {
	UserID      int64                 `json:"user_id"`
	TenantID    int64                 `json:"tenant_id"`
	Permissions []MemberPermissionDTO `json:"permissions"`
}

type MemberPermissionDTO struct {
	Permission string    `json:"permission"`
	Created    time.Time `json:"created"`
}

func NewTenantService(tenantStore tenantStore) *TenantService {
	return &TenantService{
		tenantStore: tenantStore,
	}
}

// Creates a new tenant, if a parent tenant is given it must be found and have an active state,
// otherwise ErrParentTenantNotFound is returned
func (s *TenantService) CreateNewTenant(dto TenantDTO) (TenantDTO, error) {
	if !auth.PermissionsValid(dto.Permissions) {
		return TenantDTO{}, auth.ErrGivenPermissionsNotValid
	}
	tenant := newTenantFromDto(dto)
	if tenant.ParentID != nil {
		parent, err := s.tenantStore.GetTenantById(*tenant.ParentID)
		if err != nil {
			return TenantDTO{}, err
		}
		if parent.State != Active {
			return TenantDTO{}, ErrTenantNotActive
		}
	}
	err := s.tenantStore.Create(&tenant)
	if err != nil {
		return TenantDTO{}, err
	}
	res := newTenantDtoFromTenant(tenant)
	return res, nil
}

// Sets a tenant's state to Archived
// ErrTenantNotFound is returned if the tenant is not found or the state has already been set to Archived
func (s *TenantService) ArchiveTenant(tenantID int64) error {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return err
	}
	if tenant.State == Archived {
		return ErrTenantNotActive
	}
	tenant.State = Archived
	return s.tenantStore.Update(tenant)
}

func (s *TenantService) ListTenants(filter Filter, p pagination.Request) (*pagination.Page[TenantDTO], error) {
	return s.tenantStore.List(filter, p)
}

// TODO: when a first permission is added for a user to a tenant, is that acceptable?
// Or more secure to explictely add a user to an organisation first
// TODO: should permissions have expiration date?
// Or maybe have a specific permissions that specifies their access to the tenant

// Adds the given permissions for the user_id and tenant_id combination. Errors if the permissions are not valid, if
// the tenant doesn't have the requested permissions or of the member already has the permission
func (s *TenantService) AddMemberPermissions(memberPermissions AddMemberPermissionsDTO) (MemberPermissionsAddedDTO, error) {
	if !auth.PermissionsValid(memberPermissions.Permissions) {
		return MemberPermissionsAddedDTO{}, auth.ErrGivenPermissionsNotValid
	}
	tenant, err := s.tenantStore.GetTenantById(memberPermissions.TenantID)
	if err != nil {
		return MemberPermissionsAddedDTO{}, err
	}

	// Ensure all requested permissions are actually present for the tenant
	if !tenant.ContainsAll(memberPermissions.Permissions) {
		return MemberPermissionsAddedDTO{}, auth.ErrPermissionsNotGranted
	}

	memberPerm, err := s.tenantStore.GetMemberPermissions(memberPermissions.UserID, memberPermissions.TenantID)
	if err != nil {
		return MemberPermissionsAddedDTO{}, err
	}

	// Ensure the member doesnt already have this permission
	if memberPerm.ContainsAtLeastOne(memberPermissions.Permissions) {
		return MemberPermissionsAddedDTO{}, auth.ErrAlreadyContainsPermission
	}

	// All checks are met, create the new member permissions
	permissions := memberPermissionsFromDTO(memberPermissions)
	err = s.tenantStore.CreateMemberPermissions(&permissions)
	if err != nil {
		return MemberPermissionsAddedDTO{}, err
	}
	return memberPermissionsAddedDTOFromMemberPermissions(permissions), nil
}

type tenantStore interface {
	Create(*Tenant) error
	Update(Tenant) error
	GetTenantById(id int64) (Tenant, error)
	List(Filter, pagination.Request) (*pagination.Page[TenantDTO], error)
	CreateMemberPermissions(*MemberPermissions) error
	GetMemberPermissions(userId int64, tenantId int64) (MemberPermissions, error)
}
