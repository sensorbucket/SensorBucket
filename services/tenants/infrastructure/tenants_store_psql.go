package tenantsinfra

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	pgt "github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var (
	_ tenants.TenantStore = (*PSQLTenantStore)(nil)
	_ apikeys.TenantStore = (*PSQLTenantStore)(nil)

	ErrNoRowsAffected = errors.New("no rows where updated")

	pq = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
)

type PSQLTenantStore struct {
	db *sqlx.DB
}

func NewTenantsStorePSQL(db *sqlx.DB) *PSQLTenantStore {
	return &PSQLTenantStore{
		db: db,
	}
}

type tenantsQueryPage struct {
	Created time.Time `pagination:"created,DESC"`
}

func (ts *PSQLTenantStore) GetTenantByID(id int64) (*tenants.Tenant, error) {
	tenant := tenants.Tenant{}
	q := sq.Select(
		"id, name, address, zip_code, city, chamber_of_commerce_id, headquarter_id, archive_time, state, logo, parent_tenant_id").
		From("tenants").Where(sq.Eq{"id": id})
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return nil, fmt.Errorf("could not create GetTenantByID query: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
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
			&tenant.ParentID)
		if err != nil {
			return nil, fmt.Errorf("in GetTenantByID, could not scan row: %w", err)
		}
	} else {
		return nil, tenants.ErrTenantNotFound
	}
	return &tenant, nil
}

func (ts *PSQLTenantStore) List(filter tenants.StoreFilter, r pagination.Request) (*pagination.Page[tenants.CreateTenantDTO], error) {
	var err error

	// Pagination
	cursor, err := pagination.GetCursor[tenantsQueryPage](r)
	if err != nil {
		return nil, fmt.Errorf("in List tenants, could not parse pagination cursor: %w", err)
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
	if filter.MemberID != "" {
		subQ := sq.Select("member.tenant_id").From("tenant_members member").Where(sq.Eq{"user_id": filter.MemberID})
		q = q.Where(subQ.Prefix("id IN (")).Suffix(")")
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, fmt.Errorf("in List tenants, could not apply pagination: %w", err)
	}
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return nil, fmt.Errorf("in List tenants, could not run query: %w", err)
	}
	defer rows.Close()
	list := make([]tenants.CreateTenantDTO, 0, cursor.Limit)
	for rows.Next() {
		tenant := tenants.CreateTenantDTO{}
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
			return nil, fmt.Errorf("in List tenants, could not scan row: %w", err)
		}
		list = append(list, tenant)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (ts *PSQLTenantStore) Create(tenant *tenants.Tenant) error {
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
			tenant.ParentID)
	q = q.PlaceholderFormat(sq.Dollar).Suffix("RETURNING \"id\"").RunWith(ts.db)
	return q.QueryRow().Scan(&tenant.ID)
}

func (ts *PSQLTenantStore) Update(tenant *tenants.Tenant) error {
	tx, err := ts.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("could not start database Transaction: %w", err)
	}
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
	res, err := q.PlaceholderFormat(sq.Dollar).RunWith(tx).Exec()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}
	updated, err := res.RowsAffected()
	if err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return fmt.Errorf("error getting affected rows: %w", err)
	}
	if updated != 1 {
		if updated == 0 {
			var err error = tenants.ErrTenantNotFound
			if rb := tx.Rollback(); rb != nil {
				err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
			}
			return err
		} else {
			err := fmt.Errorf("more than one row was updated")
			if rb := tx.Rollback(); rb != nil {
				err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
			}
			return err
		}
	}

	if rb := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", rb)
	}
	return nil
}

func (store *PSQLTenantStore) GetTenantHierarchyChildren(startingTenantIDs []int64) ([]tenants.Tenant, error) {
	anchor := sq.Select("t.*").From("tenants t").Where(sq.Eq{"t.id": startingTenantIDs})
	return getTenantHierarchy(store.db, anchor)
}

func (store *PSQLTenantStore) GetUserTenants(userID string) ([]tenants.Tenant, error) {
	anchor := sq.Select("t.*").From("tenant_members m").Where(sq.Eq{"user_id": userID}).LeftJoin("tenants t ON m.tenant_id = t.id")
	return getTenantHierarchy(store.db, anchor)
}

// getTenantHierarchy anchor point is a select query that must return at least an id and pare
func getTenantHierarchy(db *sqlx.DB, anchor sq.SelectBuilder) ([]tenants.Tenant, error) {
	recursive := sq.Select("t.*").From("tenants t").InnerJoin("children c ON t.parent_tenant_id = c.id")
	cte := anchor.SuffixExpr(sq.ConcatExpr(" UNION ", recursive)).Prefix("WITH RECURSIVE children AS (").Suffix(")")
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
	).From("children").PrefixExpr(cte)

	query, params, err := q.PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	row, err := db.Queryx(query, params...)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	tenantList := make([]tenants.Tenant, 0)
	for row.Next() {
		var tenant tenants.Tenant
		err = row.Scan(
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
		)
		if err != nil {
			return nil, err
		}
		tenantList = append(tenantList, tenant)
	}
	return tenantList, nil
}

func (store *PSQLTenantStore) GetMember(tenantID int64, userID string) (*tenants.Member, error) {
	var member tenants.Member
	var permissions pgt.FlatArray[auth.Permission]
	// TODO: This is a hack, it grabs the underlying PGX connection to scan the row
	// it is sort of required to scan the array of permission as that column is a postgres
	// special type
	c, err := store.db.Conn(context.Background())
	if err != nil {
		return nil, fmt.Errorf("in GetTenantMember, could not get raw db conn: %w", err)
	}
	defer c.Close()
	err = c.Raw(func(driverConn any) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("in GetTenantMember, expected driverConnection to be of type stdlib.Conn")
		}
		conn := stdlibConn.Conn()
		query, params, err := pq.Select("tenant_id", "user_id", "permissions").From("tenant_members").Where(sq.Eq{"tenant_id": tenantID, "user_id": userID}).ToSql()
		if err != nil {
			return fmt.Errorf("in GetTenantMember, could not build sql query: %w", err)
		}
		err = conn.QueryRow(context.Background(), query, params...).
			Scan(
				&member.TenantID, &member.UserID, &permissions,
			)
		if errors.Is(err, pgx.ErrNoRows) {
			return tenants.ErrTenantMemberNotFound
		} else if err != nil {
			return fmt.Errorf("error getting tenant member: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err // error context given by method above
	}
	member.Permissions = auth.Permissions(permissions)
	return &member, nil
}

func (store *PSQLTenantStore) IsMember(tenantID int64, userID string, explicit bool) (bool, error) {
	var count int64
	if explicit {
		q := sq.Select("count(*)").From("tenant_members m").Where(sq.Eq{"user_id": userID, "tenant_id": tenantID})
		err := q.PlaceholderFormat(sq.Dollar).RunWith(store.db).Scan(&count)
		if err != nil {
			return false, fmt.Errorf("in IsMember PSQLStore: %w", err)
		}
		return count > 0, nil
	}

	// Recursively create parent chain for this tenant,
	// inner join to tenant_members and count rows
	// if count > 0, then atleast one tenant in this chain has an explicit membership

	anchor := sq.Select("t.id", "t.parent_tenant_id").From("tenants t").Where(sq.Eq{"t.id": tenantID})
	recursive := sq.Select("t.id", "t.parent_tenant_id").From("tenants t").
		InnerJoin("(SELECT * FROM tenant_members WHERE user_id = ?) options ON options.tenant_id = t.id", userID).
		InnerJoin("children c ON t.id = c.parent_tenant_id")
	cte := anchor.SuffixExpr(sq.ConcatExpr(" UNION ", recursive)).Prefix("WITH RECURSIVE children AS (").Suffix(")")
	q := sq.Select("count(id)").From("children").PrefixExpr(cte)

	if err := q.PlaceholderFormat(sq.Dollar).RunWith(store.db).Scan(&count); err != nil {
		return false, fmt.Errorf("in IsMember PSQLStore: %w", err)
	}
	return count > 0, nil
}

func (store *PSQLTenantStore) GetImplicitMemberPermissions(tenantID int64, userID string) (auth.Permissions, error) {
	// This is a cool query
	// It starts at tenant with id=tenantID, then recursively finds its parent
	// and its parent's parent, etc...
	// then join this hierarchy with the tenant_members table where user_id=userID
	// at last it unnests the permissions column (which is a VARCHAR array) and makes it
	// distinct, so that we have all unique permissions from the whole tenant hierarchy
	// for this user
	var permissions auth.Permissions
	err := store.db.Select(&permissions,
		`WITH permissionList AS (
          WITH RECURSIVE hierarchy AS (
             SELECT id, parent_tenant_id FROM tenants WHERE id = $1 
             UNION
             SELECT t.id, t.parent_tenant_id FROM tenants t INNER JOIN hierarchy p ON t.id = p.parent_tenant_id
          )
          SELECT permissions FROM tenant_members member LEFT JOIN hierarchy h ON member.tenant_id = h.id 
            WHERE user_id = $2
        )
        SELECT DISTINCT unnest(permissions) FROM permissionList`,
		tenantID, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("in GetTenantMember PSQL Store, could not query member permissions: %w", err)
	}
	if err := permissions.Validate(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (s *PSQLTenantStore) SaveMember(tenantID int64, member *tenants.Member) error {
	var count int64
	if err := s.db.Get(&count, "SELECT count(*) FROM tenant_members WHERE tenant_id = $1 AND user_id = $2", tenantID, member.UserID); err != nil {
		return fmt.Errorf("error getting counting tenant member for existance: %w", err)
	}
	if count > 0 {
		return s.updateMember(tenantID, member)
	}
	return s.createMember(tenantID, member)
}

func (s *PSQLTenantStore) createMember(tenantID int64, member *tenants.Member) error {
	rows, err := pq.Insert("tenant_members").Columns("tenant_id", "user_id", "permissions").
		Values(tenantID, member.UserID, member.Permissions).RunWith(s.db).Exec()
	if err != nil {
		return fmt.Errorf("error inserting new tenant member: %w", err)
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("in CreateMember, could not get affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *PSQLTenantStore) updateMember(tenantID int64, member *tenants.Member) error {
	rows, err := pq.Update("tenant_members").Set("permissions", member.Permissions).RunWith(s.db).Exec()
	if err != nil {
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("in UpdateMember, could not get affected rows: %w", err)
	}
	if affected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}

func (s *PSQLTenantStore) RemoveMember(tenantID int64, userID string) error {
	rows, err := pq.Delete("tenant_members").Where(sq.Eq{
		"tenant_id": tenantID,
		"user_id":   userID,
	}).RunWith(s.db).Exec()
	if err != nil {
		return fmt.Errorf("in RemoveMember, could not execute query: %w", err)
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("in RemoveMember, could not get rows affected: %w", err)
	}
	if affected == 0 {
		return ErrNoRowsAffected
	}
	return nil
}
