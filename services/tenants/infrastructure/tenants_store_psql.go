package tenantsinfra

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/samber/lo"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var _ tenants.TenantStore = (*PSQLTenantStore)(nil)

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

func (ts *PSQLTenantStore) GetTenantById(id int64) (*tenants.Tenant, error) {
	tenant := tenants.Tenant{}
	q := sq.Select(
		"id, name, address, zip_code, city, chamber_of_commerce_id, headquarter_id, archive_time, state, logo, parent_tenant_id").
		From("tenants").Where(sq.Eq{"id": id})
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return nil, err
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
			return nil, err
		}
	} else {
		return nil, tenants.ErrTenantNotFound
	}
	return &tenant, nil
}

func (ts *PSQLTenantStore) List(filter tenants.Filter, r pagination.Request) (*pagination.Page[tenants.TenantDTO], error) {
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
	// Update members
	if err := updateMembers(tx, tenant); err != nil {
		if rb := tx.Rollback(); rb != nil {
			err = fmt.Errorf("rollback error %w while handling error: %w", rb, err)
		}
		return err
	}

	if rb := tx.Commit(); err != nil {
		return fmt.Errorf("commit error: %w", rb)
	}
	return nil
}

func updateMembers(tx *sqlx.Tx, tenant *tenants.Tenant) error {
	newMembers := lo.Filter(tenant.Members, func(item tenants.Member, index int) bool {
		return item.MemberID == 0
	})
	q := sq.Insert("tenant_members").Columns("tenant_id", "user_id", "permissions")
	for _, member := range newMembers {
		q = q.Values(tenant.ID, member.UserID, member.Permissions.Permissions())
	}
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(tx).Exec()
	if err != nil {
		return fmt.Errorf("error inserting new members in database: %w", err)
	}

	// Get exisitng member ids
	rows, err := sq.Select("id").From("tenant_members").Where("tenant_id = ?", tenant.ID).PlaceholderFormat(sq.Dollar).RunWith(tx).Query()
	if err != nil {
		return fmt.Errorf("could not fetch existing members from database: %w", err)
	}
	defer rows.Close()
	existingMemberIDs := []int64{}
	for rows.Next() {
		var id int64
		err := rows.Scan(&id)
		if err != nil {
			return fmt.Errorf("could not scan member id into int64: %w", err)
		}
		existingMemberIDs = append(existingMemberIDs, id)
	}

	currentMemberIDs := lo.Map(tenant.Members, func(item tenants.Member, index int) int64 { return item.MemberID })
	deletedMembers, _ := lo.Difference(existingMemberIDs, currentMemberIDs)
	_, err = sq.Delete("tenant_members").Where(sq.Eq{"id": deletedMembers}).PlaceholderFormat(sq.Dollar).RunWith(tx).Exec()
	if err != nil {
		return fmt.Errorf("could not delete removed tenants from database: %w", err)
	}

	// update test set info=tmp.info from (values (1,'new1'),(2,'new2'),(6,'new6')) as tmp (id,info) where test.id=tmp.id;
	updatedMembers := lo.Filter(tenant.Members, func(item tenants.Member, index int) bool {
		return item.IsDirty()
	})
	for _, member := range updatedMembers {
		updateQ := sq.Update("tenant_members").Set("permissions", member.Permissions.Permissions()).Where("member_id = ?", member.MemberID)
		_, err = updateQ.PlaceholderFormat(sq.Dollar).RunWith(tx).Exec()
		if err != nil {
			return fmt.Errorf("error updating member in database: %w", err)
		}
	}

	return nil
}
