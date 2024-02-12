package tenants

import (
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type TenantService struct {
	tenantStore TenantStore
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
}

func NewTenantService(tenantStore TenantStore) *TenantService {
	return &TenantService{
		tenantStore: tenantStore,
	}
}

// Creates a new tenant, if a parent tenant is given it must be found and have an active state,
// otherwise ErrParentTenantNotFound is returned
func (s *TenantService) CreateNewTenant(dto TenantDTO) (TenantDTO, error) {
	tenant := NewTenant(dto)
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

func (s *TenantService) AddTenantMember(tenantID int64, userID string, permissions auth.Permissions) error {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return err
	}
	if err := tenant.AddMember(userID); err != nil {
		return err
	}
	if err := tenant.GrantPermission(userID, permissions); err != nil {
		return err
	}
	return s.tenantStore.Update(tenant)
}

func (s *TenantService) RemoveTenantMember(tenantID int64, userID string) error {
	return ErrNotImplemented
}

func (s *TenantService) ModifyMemberPermissions(tenantID int64, userID string, permissions auth.Permissions) error {
	return ErrNotImplemented
}

type TenantStore interface {
	Create(*Tenant) error
	Update(*Tenant) error
	GetTenantById(id int64) (*Tenant, error)
	List(Filter, pagination.Request) (*pagination.Page[TenantDTO], error)
}
