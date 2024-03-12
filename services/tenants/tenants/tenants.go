package tenants

import (
	"errors"
	"net/http"
	"time"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

type State int

var (
	// Tenant States
	Unknown  State = 0
	Active   State = 1
	Archived State = 2

	// Errors
	ErrTenantNotActive      = web.NewError(http.StatusNotFound, "Tenant could not be found", "TENANT_NOT_FOUND")
	ErrTenantNotFound       = web.NewError(http.StatusNotFound, "Tenant could not be found", "TENANT_NOT_FOUND")
	ErrTenantMemberNotFound = web.NewError(http.StatusNotFound, "User is not a member of tenant", "USER_NOT_TENANT_MEMBER")
	ErrAlreadyMember        = web.NewError(http.StatusBadRequest, "User is already a member of this tennant", "USER_ALREADY_MEMBER")
	ErrNotImplemented       = errors.New("not implemented")
)

type Tenant struct {
	ID                  int64
	Name                string
	Address             string
	ZipCode             string
	City                string
	ChamberOfCommerceID *string
	HeadquarterID       *string
	ArchiveTime         *time.Duration
	State               State
	Logo                *string
	ParentID            *int64
}

func NewTenant(dto CreateTenantDTO) Tenant {
	return Tenant{
		Name:                dto.Name,
		Address:             dto.Address,
		ZipCode:             dto.ZipCode,
		City:                dto.City,
		State:               Active,
		Logo:                dto.Logo,
		ParentID:            dto.ParentID,
		ChamberOfCommerceID: dto.ChamberOfCommerceID,
		HeadquarterID:       dto.HeadquarterID,
		ArchiveTime:         dto.ArchiveTime,
	}
}

func newTenantDtoFromTenant(tenant Tenant) CreateTenantDTO {
	return CreateTenantDTO{
		ID:                  tenant.ID,
		Name:                tenant.Name,
		Address:             tenant.Address,
		ZipCode:             tenant.ZipCode,
		City:                tenant.City,
		ChamberOfCommerceID: tenant.ChamberOfCommerceID,
		HeadquarterID:       tenant.HeadquarterID,
		ArchiveTime:         tenant.ArchiveTime,
		Logo:                tenant.Logo,
		ParentID:            tenant.ParentID,
	}
}

type Member struct {
	TenantID    int64
	UserID      string
	Permissions auth.Permissions
}

func newMember(userID string) Member {
	return Member{
		UserID:      userID,
		Permissions: auth.Permissions{},
	}
}
