package tenants

import (
	"net/http"
	"time"

	"sensorbucket.nl/sensorbucket/internal/web"
)

type State int

var (
	// Tenant States
	Unknown  State = 0
	Active   State = 1
	Archived State = 2

	// Errors
	ErrTenantNotActive = web.NewError(http.StatusNotFound, "Tenant could not be found", "TENANT_NOT_FOUND")
	ErrTenantNotFound  = web.NewError(http.StatusNotFound, "Tenant could not be found", "TENANT_NOT_FOUND")
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
	Permissions         []string
}

func (t *Tenant) ContainsAll(permissions []string) bool {
	for _, requestedPerm := range permissions {
		found := false
		for _, tenantPerm := range t.Permissions {
			if requestedPerm == tenantPerm {
				found = true
			}
		}
		if !found {
			return false
		}
	}
	return true
}

type MemberPermission struct {
	ID         int64
	UserID     int64
	TenantID   int64
	Permission string
	Created    time.Time
}

type MemberPermissions []MemberPermission

func (mp MemberPermissions) ContainsAtLeastOne(permissions []string) bool {
	for _, req := range permissions {
		for _, perm := range mp {
			if req == perm.Permission {
				return true
			}
		}
	}
	return false
}

func newTenantFromDto(dto TenantDTO) Tenant {
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
		Permissions:         dto.Permissions,
	}
}

func newTenantDtoFromTenant(tenant Tenant) TenantDTO {
	return TenantDTO{
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
		Permissions:         tenant.Permissions,
	}
}

func memberPermissionsFromDTO(dto MemberPermissionsMutationDTO) MemberPermissions {
	mp := MemberPermissions{}
	for _, perm := range dto.Permissions {
		mp = append(mp, MemberPermission{
			UserID:     dto.UserID,
			TenantID:   dto.TenantID,
			Permission: perm,
		})
	}
	return mp
}

func memberPermissionsAddedDTOFromMemberPermissions(permissions []MemberPermission) MemberPermissionsAddedDTO {
	dto := MemberPermissionsAddedDTO{}
	for _, perm := range permissions {
		dto.TenantID = perm.TenantID
		dto.UserID = perm.UserID
		dto.Permissions = append(dto.Permissions, MemberPermissionDTO{
			ID:         perm.ID,
			Permission: perm.Permission,
			Created:    perm.Created,
		})
	}
	return dto
}
