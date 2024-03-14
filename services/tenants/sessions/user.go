package sessions

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"sensorbucket.nl/sensorbucket/internal/web"
)

var (
	ErrPreferenceNotSet = errors.New("preference is not set")
	ErrUserNotAMember   = web.NewError(http.StatusBadRequest, "the user is not a member of the chosen tenant and it cannot be set as preferred tenant", "ERR_USER_NOT_A_MEMBER")
)

type UserPreferenceStore interface {
	ActiveTenantID(userID string) (int64, error)
	IsUserTenantMember(userID string, tenantID int64) (bool, error)
	SetActiveTenantID(userID string, tenantID int64) error
}

type UserPreferenceService struct {
	store UserPreferenceStore
}

func NewUserPreferenceService(store UserPreferenceStore) *UserPreferenceService {
	return &UserPreferenceService{
		store: store,
	}
}

func (userPreferenceService *UserPreferenceService) ActiveTenantID(ctx context.Context, userID string) (int64, error) {
	tenantID, err := userPreferenceService.store.ActiveTenantID(userID)
	if err != nil {
		return 0, err
	}
	isMember, err := userPreferenceService.store.IsUserTenantMember(userID, tenantID)
	if err != nil {
		return 0, err
	}
	if !isMember {
		err := userPreferenceService.SetActiveTenantID(ctx, userID, 0)
		if err != nil {
			log.Printf("Tried resetting user active tenant since the user is not a member anymore, but the update failed: %v\n", err)
		}
		return 0, ErrPreferenceNotSet
	}
	return tenantID, nil
}

func (userPreferenceService *UserPreferenceService) SetActiveTenantID(ctx context.Context, userID string, tenantID int64) error {
	// tenantID 0 is a special case and unsets the active tenant, therefor membership check is not required
	if tenantID > 0 {
		isMember, err := userPreferenceService.store.IsUserTenantMember(userID, tenantID)
		if err != nil {
			return fmt.Errorf("in SetActiveTenantID PSQL Store, while validating user membership with tenant, error occured: %w", err)
		}
		if !isMember {
			return ErrUserNotAMember
		}
	}
	return userPreferenceService.store.SetActiveTenantID(userID, tenantID)
}
