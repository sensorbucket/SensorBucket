package sessions_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
)

func TestTenantActiveButUserNotMemberShouldError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	activeTenantID := int64(15)
	store := &UserPreferenceStoreMock{
		ActiveTenantIDFunc: func(userID string) (int64, error) {
			return activeTenantID, nil
		},
		IsMemberFunc: func(ctx context.Context, tenantID int64, userID string, explicit bool) (bool, error) {
			return false, nil
		},
		SetActiveTenantIDFunc: func(userID string, tenantID int64) error {
			return nil
		},
	}
	service := sessions.NewUserPreferenceService(store)

	tenantID, err := service.ActiveTenantID(ctx, userID)
	assert.ErrorIs(t, err, sessions.ErrPreferenceNotSet)
	assert.EqualValues(t, 0, tenantID)
	assert.Len(t, store.calls.ActiveTenantID, 1)
	assert.Len(t, store.calls.SetActiveTenantID, 1, "expected an update to active tenant id")
	assert.Len(t, store.calls.IsMember, 1, "expected service to validate if user is a member")
	assert.EqualValues(t, userID, store.calls.ActiveTenantID[0].UserID)
	assert.EqualValues(t, userID, store.calls.SetActiveTenantID[0].UserID)
	assert.EqualValues(t, userID, store.calls.IsMember[0].UserID)
	assert.EqualValues(t, activeTenantID, store.calls.IsMember[0].TenantID)
}

func TestSettingTenantWithoutMembershipShouldError(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New().String()
	activeTenantID := int64(15)
	store := &UserPreferenceStoreMock{
		IsMemberFunc: func(ctx context.Context, tenantID int64, userID string, explicit bool) (bool, error) {
			return false, nil
		},
		SetActiveTenantIDFunc: func(userID string, tenantID int64) error {
			return nil
		},
	}
	service := sessions.NewUserPreferenceService(store)

	err := service.SetActiveTenantIDForUser(ctx, userID, activeTenantID)
	assert.ErrorIs(t, err, sessions.ErrUserNotAMember)
	assert.Len(
		t,
		store.calls.SetActiveTenantID,
		0,
		"should not update active tenant if user is not a member",
	)
	assert.Len(t, store.calls.IsMember, 1, "expected service to validate if user is a member")
	assert.EqualValues(t, userID, store.calls.IsMember[0].UserID)
	assert.EqualValues(t, activeTenantID, store.calls.IsMember[0].TenantID)
}
