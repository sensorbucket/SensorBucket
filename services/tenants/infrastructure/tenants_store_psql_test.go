package tenantsinfra_test

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	tenantsinfra "sensorbucket.nl/sensorbucket/services/tenants/infrastructure"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

type tenant struct {
	id       int64
	name     string
	children []tenant
	members  []lo.Tuple2[string, auth.Permissions]
	parent   int64
}

func (tenant *tenant) create(t *testing.T, db *sqlx.DB, ix int) {
	// Create this tenant first
	var parent sql.NullInt64
	if tenant.parent > 0 {
		parent.Valid = true
		parent.Int64 = tenant.parent
	}
	if tenant.name == "" {
		tenant.name = fmt.Sprintf("%2X%2X", tenant.parent, tenant.id)
	}
	err := db.Get(&tenant.id,
		`INSERT INTO
            tenants (name,address,zip_code,city,state,created,parent_tenant_id)
        VALUES
            ($1, $2, $3, $4, $5, NOW(), $6)
        RETURNING id;`,
		tenant.name, "Fakestreet", "1234AB", "Fakecity", 1, parent,
	)
	if err != nil {
		require.NoError(t, err, "seeding failed")
	}
	ix++

	// Create child tenants with this as parent
	for i := range tenant.children {
		child := &tenant.children[i]
		child.parent = tenant.id
		child.create(t, db, ix)
	}

	// Create explicit memberships
	for _, member := range tenant.members {
		user, permissions := member.Unpack()
		pv := pgtype.FlatArray[auth.Permission](permissions)
		_, err := db.Exec(`
            INSERT INTO tenant_members (user_id, tenant_id, permissions)
            VALUES
            ($1, $2, $3);
        `, user, tenant.id, pv)
		if err != nil {
			require.NoError(t, err, "seeding failed")
		}
	}
}

func TestTenantSeedingFunctionality(t *testing.T) {
	db := createPostgresServer(t, false)
	store := tenantsinfra.NewTenantsStorePSQL(db)
	hierarchy := &tenant{
		children: []tenant{
			{
				children: []tenant{},
				members:  []lo.Tuple2[string, auth.Permissions]{},
			},
			{
				children: []tenant{},
				members:  []lo.Tuple2[string, auth.Permissions]{},
			},
			{
				children: []tenant{
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
				},
				members: []lo.Tuple2[string, auth.Permissions]{},
			},
		},
		members: []lo.Tuple2[string, auth.Permissions]{},
	}
	hierarchy.create(t, db, 1)

	tenant, err := store.GetTenantByID(hierarchy.id)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)

	page, err := store.List(tenants.StoreFilter{}, pagination.Request{})
	assert.NoError(t, err)
	assert.Len(t, page.Data, 6)
}

func TestGetTenantHierarchy(t *testing.T) {
	db := createPostgresServer(t, false)
	store := tenantsinfra.NewTenantsStorePSQL(db)
	hierarchy := &tenant{
		members: []lo.Tuple2[string, auth.Permissions]{},
		children: []tenant{
			{
				name:    "otherparent",
				members: []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{
					{
						name:     "otherchild",
						members:  []lo.Tuple2[string, auth.Permissions]{},
						children: []tenant{},
					},
				},
			},
			{
				members:  []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{},
			},
			{
				name: "parent",
				members: []lo.Tuple2[string, auth.Permissions]{
					{A: userID, B: auth.Permissions{}},
				},
				children: []tenant{
					{
						name:     "child1",
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
					{
						name:     "child2",
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
				},
			},
		},
	}
	hierarchy.create(t, db, 1)

	t.Run("", func(t *testing.T) {
		tenantList, err := store.GetTenantHierarchyChildren(
			[]int64{hierarchy.children[2].id},
		)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 3)
		names := lo.Map(tenantList, func(item tenants.Tenant, index int) string {
			return item.Name
		})
		assert.ElementsMatch(t, []string{"parent", "child1", "child2"}, names)
	})

	t.Run("", func(t *testing.T) {
		tenantList, err := store.GetTenantHierarchyChildren(
			[]int64{hierarchy.children[2].children[0].id},
		)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 1)
		names := lo.Map(tenantList, func(item tenants.Tenant, index int) string {
			return item.Name
		})
		assert.ElementsMatch(t, []string{"child1"}, names)
	})

	t.Run("multiple ids to start", func(t *testing.T) {
		tenantList, err := store.GetTenantHierarchyChildren(
			[]int64{
				hierarchy.children[0].id,
				hierarchy.children[2].id,
			},
		)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 5)
		names := lo.Map(tenantList, func(item tenants.Tenant, index int) string {
			return item.Name
		})
		assert.ElementsMatch(t, []string{
			"parent", "otherparent", "child1", "child2", "otherchild",
		}, names)
	})

	t.Run("multiple ids in same chain", func(t *testing.T) {
		tenantList, err := store.GetTenantHierarchyChildren(
			[]int64{
				hierarchy.children[2].id,
				hierarchy.children[2].children[0].id,
			},
		)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 3)
		names := lo.Map(tenantList, func(item tenants.Tenant, index int) string {
			return item.Name
		})
		assert.ElementsMatch(t, []string{"parent", "child1", "child2"}, names)
	})

	t.Run("root should return all tenants", func(t *testing.T) {
		tenantList, err := store.GetTenantHierarchyChildren(
			[]int64{hierarchy.id},
		)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 7)
	})
}

func TestIsMember(t *testing.T) {
	db := createPostgresServer(t, false)
	store := tenantsinfra.NewTenantsStorePSQL(db)
	hierarchy := &tenant{
		members: []lo.Tuple2[string, auth.Permissions]{},
		children: []tenant{
			{
				members: []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{
					{
						members: []lo.Tuple2[string, auth.Permissions]{
							{A: userID, B: auth.Permissions{}},
						},
						children: []tenant{},
					},
				},
			},
			{
				members:  []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{},
			},
			{
				members: []lo.Tuple2[string, auth.Permissions]{
					{A: userID, B: auth.Permissions{}},
				},
				children: []tenant{
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
				},
			},
		},
	}
	hierarchy.create(t, db, 1)

	t.Run("explicit member", func(t *testing.T) {
		member, err := store.IsMember(hierarchy.children[0].children[0].id, userID, false)
		require.NoError(t, err)
		assert.True(t, member)
	})

	t.Run("implicit member", func(t *testing.T) {
		member, err := store.IsMember(hierarchy.children[2].children[0].id, userID, false)
		require.NoError(t, err)
		assert.True(t, member)
	})

	t.Run("not a member", func(t *testing.T) {
		member, err := store.IsMember(hierarchy.children[0].id, userID, false)
		require.NoError(t, err)
		assert.True(t, member)
	})
}

func TestGetUserTenants(t *testing.T) {
	db := createPostgresServer(t, false)
	store := tenantsinfra.NewTenantsStorePSQL(db)
	hierarchy := &tenant{
		members: []lo.Tuple2[string, auth.Permissions]{},
		children: []tenant{
			{
				members: []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{
					{
						members: []lo.Tuple2[string, auth.Permissions]{},
						children: []tenant{
							{
								members: []lo.Tuple2[string, auth.Permissions]{
									{A: userID, B: auth.Permissions{}},
								},
							},
						},
					},
				},
			},
			{
				members:  []lo.Tuple2[string, auth.Permissions]{},
				children: []tenant{},
			},
			{
				members: []lo.Tuple2[string, auth.Permissions]{
					{A: userID, B: auth.Permissions{}},
				},
				children: []tenant{
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
					{
						children: []tenant{},
						members:  []lo.Tuple2[string, auth.Permissions]{},
					},
				},
			},
		},
	}
	hierarchy.create(t, db, 1)

	t.Run("existing user", func(t *testing.T) {
		tenantList, err := store.GetUserTenants(userID)
		assert.NoError(t, err)
		assert.Len(t, tenantList, 4)
	})
	t.Run("non existing user", func(t *testing.T) {
		tenantList, err := store.GetUserTenants("00000000-0000-0000-0000-000000000000")
		assert.NoError(t, err)
		assert.Len(t, tenantList, 0)
	})
}

func TestGetPermissions(t *testing.T) {
	db := createPostgresServer(t, false)
	store := tenantsinfra.NewTenantsStorePSQL(db)

	// ROOT TENANT
	hierarchy := &tenant{
		members: []lo.Tuple2[string, auth.Permissions]{
			{A: userID, B: auth.Permissions{auth.WRITE_DEVICES}},
		},
		children: []tenant{
			// ROOT->CHILD_1
			{
				members: []lo.Tuple2[string, auth.Permissions]{
					{A: userID, B: auth.Permissions{auth.READ_DEVICES}},
				},
				children: []tenant{
					// ROOT->CHILD_1->CHILD_1
					{
						members: []lo.Tuple2[string, auth.Permissions]{},
						children: []tenant{
							{},
						},
					},
				},
			},
		},
	}
	hierarchy.create(t, db, 1)
	isolatedTenant := &tenant{
		members: []lo.Tuple2[string, auth.Permissions]{
			{A: userID, B: auth.Permissions{auth.WRITE_API_KEYS}},
		},
		children: []tenant{},
	}
	isolatedTenant.create(t, db, 10)

	t.Run("Get explicit permissions (no parent tenant)", func(t *testing.T) {
		permissions, err := store.GetImplicitMemberPermissions(hierarchy.id, userID)
		assert.NoError(t, err)
		assert.ElementsMatch(t, auth.Permissions{auth.WRITE_DEVICES}, permissions)
	})
	t.Run("Get permissions and inherit parent tenant permissions", func(t *testing.T) {
		permissions, err := store.GetImplicitMemberPermissions(hierarchy.children[0].id, userID)
		assert.NoError(t, err)
		assert.ElementsMatch(t, auth.Permissions{auth.WRITE_DEVICES, auth.READ_DEVICES}, permissions)
	})
	t.Run("Get isolated tenant permissions", func(t *testing.T) {
		permissions, err := store.GetImplicitMemberPermissions(isolatedTenant.id, userID)
		assert.NoError(t, err)
		assert.ElementsMatch(t, auth.Permissions{auth.WRITE_API_KEYS}, permissions)
	})
}
