package tenantsinfra

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

type TenantsStore struct {
	db *sqlx.DB
}

func NewTenantsStorePSQL(db *sqlx.DB) *TenantsStore {
	return &TenantsStore{
		db: db,
	}
}

type tenantsQueryPage struct {
	Created time.Time `pagination:"tenants.created,DESC"`
}

func (ts *TenantsStore) GetTenantById(id int64) (tenants.Tenant, error) {
	tenant := tenants.Tenant{}
	q := sq.Select(
		"tenants.id",
		"tenants.name",
		"tenants.address",
		"tenants.zip_code",
		"tenants.city",
		"tenants.chamber_of_commerce_id",
		"tenants.headquarter_id",
		"tenants.archive_time",
		"tenants.state",
		"tenants.logo",
		"tenants.parent_tenant_id",
		"tenant_permissions.permission").
		From("tenants").
		Join("tenant_permissions on tenants.id = tenant_permissions.tenant_id").
		Where(sq.Eq{"tenants.id": id})
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return tenants.Tenant{}, err
	}
	for rows.Next() {
		permission := ""
		err = rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Address,
			&tenant.ZipCode,
			&tenant.City,
			&tenant.ChamberOfCommerceID,
			&tenant.HeadquarterID,
			&tenant.ArchiveTime,
			&tenant.State,
			&tenant.Logo,
			&tenant.ParentID,
			&permission)
		if err != nil {
			return tenants.Tenant{}, err
		}
		tenant.Permissions = append(tenant.Permissions, permission)
	}
	if tenant.ID == 0 {
		return tenants.Tenant{}, tenants.ErrTenantNotFound
	}
	return tenant, nil
}

func (ts *TenantsStore) List(filter tenants.Filter, r pagination.Request) (*pagination.Page[tenants.TenantDTO], error) {
	var err error

	// Pagination
	cursor, err := pagination.GetCursor[tenantsQueryPage](r)
	if err != nil {
		return nil, err
	}
	q := sq.Select(
		"id",
		"name",
		"address",
		"zip_code",
		"city",
		"chamber_of_commerce_id",
		"headquarter_id",
		"archive_time",
		"logo",
		"parent_tenant_id",
	).From("tenants")
	if len(filter.Name) > 0 {
		q = q.Where(sq.Eq{"name": filter.Name})
	}
	if len(filter.State) > 0 {
		q = q.Where(sq.Eq{"state": filter.State})
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]tenants.TenantDTO, 0, cursor.Limit)
	for rows.Next() {
		tenant := tenants.TenantDTO{}
		err = rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.Address,
			&tenant.ZipCode,
			&tenant.City,
			&tenant.ChamberOfCommerceID,
			&tenant.HeadquarterID,
			&tenant.ArchiveTime,
			&tenant.Logo,
			&tenant.ParentID,
			&cursor.Columns.Created,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, tenant)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (ts *TenantsStore) Create(tenant *tenants.Tenant) error {

	// Each permission for the tenant is stored as a seperate record
	// Create the insert query where each permission is inserted as a seperate record
	tenantPermissionsQ := sq.Insert("tenant_permissions").Columns("permission", "tenant_id")
	for _, permission := range tenant.Permissions {
		tenantPermissionsQ = tenantPermissionsQ.Values(permission, sq.Select("id").From("new_tenant").Prefix("(").Suffix(")"))
	}
	tenantPermissionsQ = tenantPermissionsQ.Suffix("RETURNING \"tenant_id\"")

	q := sq.Insert("tenants").
		Columns(
			"name",
			"address",
			"zip_code",
			"city",
			"chamber_of_commerce_id",
			"headquarter_id",
			"archive_time",
			"state",
			"logo",
			"created",
			"parent_tenant_id").
		Values(
			tenant.Name,
			tenant.Address,
			tenant.ZipCode,
			tenant.City,
			tenant.ChamberOfCommerceID,
			tenant.HeadquarterID,
			tenant.ArchiveTime,
			tenant.State,
			tenant.Logo,
			time.Now().UTC(),
			tenant.ParentID).

		// Return id so the follow up permission query can use the new_tenant.id to add the permission correctly
		Prefix("WITH new_tenant AS (").
		Suffix("RETURNING \"id\")").

		// Run the permission query along with the insert new tenant query
		SuffixExpr(tenantPermissionsQ).
		PlaceholderFormat(sq.Dollar).RunWith(ts.db)

	return q.QueryRow().Scan(&tenant.ID)
}

// TODO: update, how to handle permissions?
func (ts *TenantsStore) Update(tenant tenants.Tenant) error {
	q := sq.Update("tenants").
		Set("name", tenant.Name).
		Set("address", tenant.Address).
		Set("zip_code", tenant.ZipCode).
		Set("city", tenant.City).
		Set("chamber_of_commerce_id", tenant.ChamberOfCommerceID).
		Set("headquarter_id", tenant.HeadquarterID).
		Set("archive_time", tenant.ArchiveTime).
		Set("state", tenant.State).
		Set("logo", tenant.Logo).
		Set("parent_tenant_id", tenant.ParentID).
		Where(sq.Eq{"id": tenant.ID})
	res, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Exec()
	if err != nil {
		return err
	}
	updated, err := res.RowsAffected()
	if updated != 1 {
		if updated == 0 {
			return tenants.ErrTenantNotFound
		} else {
			return fmt.Errorf("more than one row was updated")
		}
	}
	return nil
}

func (ts *TenantsStore) GetMemberPermissions(userId int64, tenantId int64) (tenants.MemberPermissions, error) {
	q := sq.
		Select(
			"member_permissions.id",
			"member_permissions.user_id",
			"tenant_permissions.tenant_id",
			"member_permissions.created",
			"tenant_permissions.permission").
		From("member_permissions").
		Join("tenant_permissions on permission_id = tenant_permissions.id").
		Where(sq.Eq{"member_permissions.user_id": userId, "tenant_permissions.tenant_id": tenantId})

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return nil, err
	}
	list := tenants.MemberPermissions{}
	for rows.Next() {
		perm := tenants.MemberPermission{}
		err = rows.Scan(
			&perm.ID,
			&perm.UserID,
			&perm.TenantID,
			&perm.Created,
			&perm.Permission,
		)
		if err != nil {
			return nil, err
		}
		list = append(list, perm)
	}
	return list, nil
}

func (ts *TenantsStore) CreateMemberPermissions(memberPermissions *tenants.MemberPermissions) error {
	q := sq.Insert("member_permissions").Columns(
		"user_id",
		"created",
		"permission_id",
	)
	for _, perm := range *memberPermissions {
		q = q.Values(
			perm.UserID,
			time.Now().UTC(),

			// Retrieve the correct tenant permission by the permission string and tenant id
			sq.Select("id").From("tenant_permissions").
				Where(sq.Eq{"permission": perm.Permission, "tenant_id": perm.TenantID}).
				Prefix("(").Suffix(")"))
	}
	q = q.Suffix("RETURNING id, created")
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return err
	}
	insertedPermissions := tenants.MemberPermissions{}
	i := 0
	for rows.Next() {
		if i > len(*memberPermissions)-1 {
			return fmt.Errorf("some permissions could not be added")
		}
		permissions := *memberPermissions
		perm := permissions[i]
		rows.Scan(
			&perm.ID,
			&perm.Created)
		insertedPermissions = append(insertedPermissions, perm)
		i++
	}
	*memberPermissions = insertedPermissions
	return nil
}
