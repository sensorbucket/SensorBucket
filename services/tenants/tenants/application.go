package tenants

import (
	"errors"
	"fmt"
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
)

var (
	ErrParentTenantNotFound = fmt.Errorf("Parent tenant could not be found")
	ErrTenantNotFound       = fmt.Errorf("Tenant could not be found")
)

func NewTenantService(tenantStore tenantStore) *service {
	return &service{
		tenantStore: tenantStore,
	}
}

func (s *service) GetTenantById(tenantID int64) (TenantDTO, error) {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return TenantDTO{}, err
	}
	res := newTenantDtoFromTenant(tenant)
	return res, nil
}

// Creates a new tenant, if a parent tenant is given it must be found and have an active state,
// otherwise ErrParentTenantNotFound is returned
func (s *service) CreateNewTenant(dto TenantDTO) (TenantDTO, error) {
	tenant := newTenantFromDto(dto)
	if tenant.ParentID != nil {
		parent, err := s.tenantStore.GetTenantById(*tenant.ParentID)
		if err != nil {
			if errors.Is(err, ErrTenantNotFound) {
				return TenantDTO{}, ErrParentTenantNotFound
			} else {
				return TenantDTO{}, err
			}
		}
		if parent.State != Active {
			return TenantDTO{}, ErrParentTenantNotFound
		}
	}
	err := s.tenantStore.Create(tenant)
	if err != nil {
		return TenantDTO{}, err
	}
	res := newTenantDtoFromTenant(tenant)
	return res, nil
}

// Sets a tenant's state to Archived
// ErrTenantNotFound is returned if the tenant is not found or the state has already been set to Archived
func (s *service) ArchiveTenant(tenantID int64) error {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return err
	}
	if tenant.State == Archived {
		return ErrTenantNotFound
	}
	tenant.State = Archived
	return s.tenantStore.Update(tenant)
}

func (s *service) ListTenants(filter Filter, p pagination.Request) (*pagination.Page[TenantDTO], error) {
	return s.tenantStore.List(filter, p)
}

type Filter struct {
	Name  []string `json:"name"`
	State []State  `json:"state"`
}

type TenantDTO struct {
	Name                string         `json:"name"`
	Address             string         `json:"address"`
	ZipCode             string         `json:"zip_code"`
	City                string         `json:"city"`
	ChamberOfCommerceID string         `json:"chamber_of_commerce_id"`
	HeadquarterID       string         `json:"headquarter_id"`
	ArchiveTime         *time.Duration `json:"archive_time"`
	Logo                *string        `json:"logo"`
	ParentID            *int64         `json:"parent_tenant_id"`
}

type service struct {
	tenantStore tenantStore
}
type tenantStore interface {
	Create(Tenant) error
	Update(Tenant) error
	GetTenantById(id int64) (Tenant, error)
	List(Filter, pagination.Request) (*pagination.Page[TenantDTO], error)
}
