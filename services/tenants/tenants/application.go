package tenants

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var (
	ErrUserNotValidated = web.NewError(http.StatusBadRequest, "Could not validate provided user ID", "ERR_USER_NOT_VALIDATED")
	ErrSessionInvalid   = web.NewError(http.StatusBadRequest, "Invalid authentication session", "ERR_SESSION_INVALID")
)

type TenantService struct {
	tenantStore   TenantStore
	userValidator UserValidator
}

type Filter struct {
	Name     []string `url:"name"`
	State    []State  `url:"state"`
	IsMember bool     `url:"is_member"`
}

type CreateTenantDTO struct {
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

func NewTenantService(tenantStore TenantStore, userValidator UserValidator) *TenantService {
	return &TenantService{
		tenantStore:   tenantStore,
		userValidator: userValidator,
	}
}

// Creates a new tenant, if a parent tenant is given it must be found and have an active state,
// otherwise ErrParentTenantNotFound is returned
func (s *TenantService) CreateNewTenant(ctx context.Context, dto CreateTenantDTO) (CreateTenantDTO, error) {
	tenant := NewTenant(dto)
	if tenant.ParentID != nil {
		parent, err := s.tenantStore.GetTenantByID(*tenant.ParentID)
		if err != nil {
			return CreateTenantDTO{}, err
		}
		if parent.State != Active {
			return CreateTenantDTO{}, ErrTenantNotActive
		}
	}
	err := s.tenantStore.Create(&tenant)
	if err != nil {
		return CreateTenantDTO{}, err
	}
	res := newTenantDtoFromTenant(tenant)
	return res, nil
}

func (s *TenantService) GetTenantByID(ctx context.Context, id int64) (*Tenant, error) {
	tenant, err := s.tenantStore.GetTenantByID(id)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

// Sets a tenant's state to Archived
// ErrTenantNotFound is returned if the tenant is not found or the state has already been set to Archived
func (s *TenantService) ArchiveTenant(ctx context.Context, tenantID int64) error {
	tenant, err := s.tenantStore.GetTenantByID(tenantID)
	if err != nil {
		return err
	}
	if tenant.State == Archived {
		return ErrTenantNotActive
	}
	tenant.State = Archived
	return s.tenantStore.Update(tenant)
}

type StoreFilter struct {
	MemberID string
	State    []State
	Name     []string
}

func (s *TenantService) ListTenants(ctx context.Context, filter Filter, p pagination.Request) (*pagination.Page[CreateTenantDTO], error) {
	var storeFilter StoreFilter
	storeFilter.State = filter.State
	storeFilter.Name = filter.Name
	if filter.IsMember {
		userID, err := auth.GetUser(ctx)
		if err != nil {
			return nil, fmt.Errorf("%w: must be authenticated as a user to use the 'IsMember' filter", ErrSessionInvalid)
		}
		storeFilter.MemberID = userID
	}
	return s.tenantStore.List(storeFilter, p)
}

func (s *TenantService) AddTenantMember(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
	if err := auth.Permissions(permissions).Validate(); err != nil {
		return err
	}
	t, err := s.tenantStore.GetTenantByID(tenantID)
	if err != nil {
		return err
	}

	// Validate that we get a NotFound error so that we know the user is not yet a member
	_, err = s.tenantStore.GetMember(t.ID, userID)
	if err == nil {
		return ErrAlreadyMember
	} else if !errors.Is(err, ErrTenantMemberNotFound) {
		return fmt.Errorf("#AddTenantMember: could not getTenantMember to verify existence: %w", err)
	}

	// Validate that the given user exists and can be added to the tenant
	if err := s.userValidator.UserByIDExists(ctx, tenantID, userID); err != nil {
		return fmt.Errorf("in AddTenantMember: %w: %w", ErrUserNotValidated, err)
	}

	member := newMember(userID)
	member.Permissions = permissions
	if err := s.tenantStore.SaveMember(t.ID, &member); err != nil {
		return err
	}

	return nil
}

func (s *TenantService) UpdateTenantMember(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
	if err := auth.Permissions(permissions).Validate(); err != nil {
		return err
	}
	member, err := s.tenantStore.GetMember(tenantID, userID)
	if err != nil {
		return fmt.Errorf("in UpdateTenantMember, could not get Tenant Member: %w", err)
	}
	// Validate that the given user exists and can be added to the tenant
	if err := s.userValidator.UserByIDExists(ctx, tenantID, userID); err != nil {
		return fmt.Errorf("in AddTenantMember: %w: %w", ErrUserNotValidated, err)
	}

	member.Permissions = permissions
	if err := s.tenantStore.SaveMember(tenantID, member); err != nil {
		return err
	}

	return nil
}

func (s *TenantService) RemoveTenantMember(ctx context.Context, tenantID int64, userID string) error {
	_, err := s.tenantStore.GetMember(tenantID, userID)
	if err != nil {
		return err
	}

	return s.tenantStore.RemoveMember(tenantID, userID)
}

func (s *TenantService) ModifyMemberPermissions(ctx context.Context, tenantID int64, userID string, permissions auth.Permissions) error {
	if err := auth.Permissions(permissions).Validate(); err != nil {
		return err
	}
	_, err := s.tenantStore.GetTenantByID(tenantID)
	if err != nil {
		return err
	}
	//if tenant.State != tenants.Active {
	//	return ErrTenantNotActive
	//}
	member, err := s.tenantStore.GetMember(tenantID, userID)
	if err != nil {
		return err
	}
	member.Permissions = permissions
	if err := s.tenantStore.SaveMember(tenantID, member); err != nil {
		return err
	}
	return nil
}

// GetMemberPermissions returns the total permission set a user has for this tenant,
// this also inherits permissions from parent tenants where the user is a member of
// Returns an error if the user is not a member
func (s *TenantService) GetMemberPermissions(ctx context.Context, tenantID int64, userID string) (auth.Permissions, error) {
	isMember, err := s.tenantStore.IsMember(tenantID, userID, false)
	if err != nil {
		return nil, fmt.Errorf("in GetMemberPermissions: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("in GetMemberPermissions: %w", ErrTenantMemberNotFound)
	}

	permissions, err := s.tenantStore.GetImplicitMemberPermissions(tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("in GetMemberPermissions, failed to get member: %w", err)
	}
	return permissions, nil
}

func (s *TenantService) GetUserTenants(ctx context.Context, userID string) ([]Tenant, error) {
	tenants, err := s.tenantStore.GetUserTenants(userID)
	if err != nil {
		return nil, fmt.Errorf("in GetUserTenants, failed to GetUserTenants: %w", err)
	}
	return tenants, nil
}

type TenantStore interface {
	Create(*Tenant) error
	Update(*Tenant) error
	GetTenantByID(id int64) (*Tenant, error)
	GetMember(tenantID int64, userID string) (*Member, error)
	GetImplicitMemberPermissions(tenantID int64, userID string) (auth.Permissions, error)
	SaveMember(tenantID int64, member *Member) error
	RemoveMember(tenantID int64, userID string) error
	List(StoreFilter, pagination.Request) (*pagination.Page[CreateTenantDTO], error)
	GetUserTenants(userID string) ([]Tenant, error)
	IsMember(tenantID int64, userID string, explicit bool) (bool, error)
}

type UserValidator interface {
	UserByIDExists(ctx context.Context, tenantID int64, userID string) error
}
