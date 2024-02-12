package tenants

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/samber/lo"

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

	Members []Member
}

func NewTenant(dto TenantDTO) Tenant {
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
		Members:             []Member{},
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
	}
}

func (t *Tenant) AddMember(userID string) error {
	if _, err := t.GetMember(userID); !errors.Is(err, ErrTenantMemberNotFound) {
		if err == nil {
			return ErrAlreadyMember
		}
		return err
	}
	t.Members = append(t.Members, newMember(userID))
	return nil
}

func (t *Tenant) RemoveMember(userID string) error {
	_, ix, ok := lo.FindIndexOf(t.Members, func(item Member) bool {
		return item.UserID == userID
	})
	if !ok {
		return ErrTenantMemberNotFound
	}
	lastIX := len(t.Members) - 1
	t.Members[ix] = t.Members[lastIX]
	t.Members = t.Members[:lastIX]
	return nil
}

func (t *Tenant) GrantPermission(userID string, permissions auth.PermissionSet) error {
	_, ix, ok := lo.FindIndexOf(t.Members, func(item Member) bool {
		return item.UserID == userID
	})
	if !ok {
		return fmt.Errorf("cannot get user for granting permission: %w", ErrTenantNotFound)
	}
	member := &t.Members[ix]
	member.Permissions = append(member.Permissions, permissions)
	member.dirty = true
	return nil
}

func (t *Tenant) RevokePermission(userID string, revokedPermissions auth.PermissionSet) error {
	member, ix, ok := lo.FindIndexOf(t.Members, func(item Member) bool {
		return item.UserID == userID
	})
	if !ok {
		return fmt.Errorf("cannot get user for revoking permission: %w", ErrTenantNotFound)
	}
	memberPermissions := member.Permissions.Permissions()
	memberPermissions = lo.Reject(memberPermissions, func(memberPermission auth.Permission, index int) bool {
		return lo.ContainsBy(revokedPermissions.Permissions(), func(revokedPermission auth.Permission) bool {
			return memberPermission.String() == revokedPermission.String()
		})
	})
	member.Permissions = lo.Map(memberPermissions, func(item auth.Permission, _ int) auth.PermissionSet {
		return item
	})
	member.dirty = true
	t.Members[ix] = member
	return nil
}

func (t *Tenant) GetMember(userID string) (Member, error) {
	member, ok := lo.Find(t.Members, func(item Member) bool { return item.UserID == userID })
	if !ok {
		return member, ErrTenantMemberNotFound
	}
	return member, nil
}

type Member struct {
	MemberID    int64
	UserID      string
	Permissions auth.Permissions

	dirty bool
}

func (m Member) IsDirty() bool {
	return m.dirty
}

func newMember(userID string) Member {
	return Member{
		UserID:      userID,
		Permissions: auth.Permissions{},
	}
}
