package sessions

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var (
	ErrPreferenceNotSet = errors.New("preference is not set")
	ErrUserNotAMember   = web.NewError(
		http.StatusBadRequest,
		"the user is not a member of the chosen tenant and it cannot be set as preferred tenant",
		"ERR_USER_NOT_A_MEMBER",
	)
)

type UserPreferenceStore interface {
	ActiveTenantID(userID string) (int64, error)
	SetActiveTenantID(userID string, tenantID int64) error
}

type UserPreferenceService struct {
	store       UserPreferenceStore
	tenantStore tenants.TenantStore
}

func NewUserPreferenceService(
	store UserPreferenceStore,
	tenantStore tenants.TenantStore,
) *UserPreferenceService {
	return &UserPreferenceService{
		store:       store,
		tenantStore: tenantStore,
	}
}

// ActiveTenantID returns the user's preferred tentant. This also validates that the user is a member, if the user
// isn't a member its preferred tenant will be set to 0 and the ErrPreferenceNotSet error will be returned and the
// caller is responsible for the next actions.
func (userPreferenceService *UserPreferenceService) ActiveTenantID(
	ctx context.Context,
	userID string,
) (int64, error) {
	tenantID, err := userPreferenceService.store.ActiveTenantID(userID)
	if err != nil {
		return 0, err
	}

	isMember, err := userPreferenceService.tenantStore.IsMember(ctx, tenantID, userID, false)
	if err != nil {
		return 0, err
	}
	if !isMember {
		log.Printf(
			"user (%s) is not a member of preferred tenant (%d), will remove preference\n",
			userID,
			tenantID,
		)

		if err := userPreferenceService.store.SetActiveTenantID(userID, 0); err != nil {
			log.Printf(
				"error: User is not a member of prefered tenant anymore, but falling back to a tenant it is a member of was not possible: %s\n",
				err.Error(),
			)
		}

		return 0, ErrPreferenceNotSet
	}

	// If we reach here, tenantID is valid and user is a member.
	return tenantID, nil
}

// SetUserPreferedTenantToFallback will set the user's prefered tenant to one it is a member of.
func (service *UserPreferenceService) SetUserPreferedTenantToFallback(
	ctx context.Context,
	userID string,
) (int64, error) {
	tenantID, err := service.getFallbackTenantID(ctx, userID)
	if err != nil {
		return 0, err
	}

	if err := service.SetActiveTenantIDForUser(ctx, userID, tenantID); err != nil {
		return 0, err
	}

	return tenantID, nil
}

func (service *UserPreferenceService) getFallbackTenantID(
	ctx context.Context,
	userID string,
) (int64, error) {
	tenants, err := service.tenantStore.List(
		ctx,
		tenants.StoreFilter{MemberID: userID, State: []tenants.State{tenants.Active}},
		pagination.Request{Limit: 1},
	)
	if err != nil {
		return 0, err
	}
	if len(tenants.Data) == 0 {
		return 0, errors.New("no tenant to fallback to")
	}

	return tenants.Data[0].ID, nil
}

func (userPreferenceService *UserPreferenceService) SetActiveTenantIDForUser(
	ctx context.Context,
	userID string,
	tenantID int64,
) error {
	if tenantID == 0 {
		log.Printf("requesting fallback tenant for user (%s)\n", userID)
		fallbackTenantID, err := userPreferenceService.getFallbackTenantID(ctx, userID)
		if err != nil {
			return err
		}
		tenantID = fallbackTenantID
		log.Printf("fallback tenant for user (%s) is %d\n", userID, tenantID)
	}

	// tenantID 0 is a special case and unsets the active tenant, therefor membership check is not required
	if tenantID > 0 {
		isMember, err := userPreferenceService.tenantStore.IsMember(ctx, tenantID, userID, false)
		if err != nil {
			return fmt.Errorf(
				"in SetActiveTenantID PSQL Store, while validating user membership with tenant, error occured: %w",
				err,
			)
		}
		if !isMember {
			return ErrUserNotAMember
		}
	}
	return userPreferenceService.store.SetActiveTenantID(userID, tenantID)
}

func (userPreferenceService *UserPreferenceService) SetActiveTenantID(
	ctx context.Context,
	tenantID int64,
) error {
	userID, err := auth.GetUser(ctx)
	if err != nil {
		return fmt.Errorf("in SetActiveTenantID, error getting userID from context: %w", err)
	}
	return userPreferenceService.SetActiveTenantIDForUser(ctx, userID, tenantID)
}
