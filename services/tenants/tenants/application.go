package tenants

import (
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

func (s *service) GetTenantById(tenantID int64) (*TenantDTO, error) {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return nil, err
	}
	if tenant == nil {
		return nil, ErrTenantNotFound
	}
	res := newTenantDtoFromTenant(*tenant)
	return &res, nil
}

func (s *service) CreateNewTenant(dto TenantDTO) (*TenantDTO, error) {
	tenant := newTenantFromDto(dto)
	if tenant.ParentID != nil {
		parent, err := s.GetTenantById(*tenant.ParentID)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			return nil, ErrParentTenantNotFound
		}
	}
	err := s.tenantStore.Create(&tenant)
	if err != nil {
		return nil, err
	}
	res := newTenantDtoFromTenant(tenant)
	return &res, nil
}

// TODO: discuss: how should not found errors be done? Return nil pointer or error to http layer??
func (s *service) ArchiveTenant(tenantID int64) error {
	tenant, err := s.tenantStore.GetTenantById(tenantID)
	if err != nil {
		return err
	}
	if tenant == nil {
		return ErrTenantNotFound
	}
	tenant.State = Archived
	return s.tenantStore.Update(tenant)
}

func (s *service) ListTenants(filter Filter, p pagination.Request) (*pagination.Page[TenantDTO], error) {
	return s.tenantStore.List(filter, p)
}

type Filter struct {
	Name []string `json:"name"`
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
	Create(*Tenant) error
	Update(*Tenant) error
	GetTenantById(id int64) (*Tenant, error)
	List(Filter, pagination.Request) (*pagination.Page[TenantDTO], error)
}
