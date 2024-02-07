package tenants_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestAddMemberToTenantWithPermissions(t *testing.T) {
	tenant := tenants.NewTenant(tenants.TenantDTO{
		Name: "TestingTenant",
	})
	userID := "123123"

	require.NoError(t, tenant.AddMember(userID), "could not add member")
	require.NoError(t, tenant.GrantPermission(userID, auth.READ_DEVICES), "could not grant permissions")
	require.NoError(t, tenant.GrantPermission(userID, auth.WRITE_DEVICES), "could not grant permissions")
	member, err := tenant.GetMember(userID)
	require.NoError(t, err)
	assert.Equal(t, auth.Permissions{auth.READ_DEVICES, auth.WRITE_DEVICES}, member.Permissions)
	require.NoError(t, tenant.RevokePermission(userID, auth.WRITE_DEVICES), "could not revoke permission")
	assert.Equal(t, auth.Permissions{auth.READ_DEVICES}, member.Permissions)
	require.NoError(t, tenant.RemoveMember(userID), "could not remove member")
	_, err = tenant.GetMember(userID)
	require.ErrorIs(t, err, tenants.ErrTenantMemberNotFound, "expected member not found error")
	require.NoError(t, tenant.AddMember(userID), "could not add member")
	member, err = tenant.GetMember(userID)
	require.NoError(t, err)
	assert.Equal(t, auth.Permissions{}, member.Permissions, "member should not have any permissions left after removal")
	require.ErrorIs(t, tenant.AddMember(userID), tenants.ErrAlreadyMember, "should not be able to add same user twice")
}
