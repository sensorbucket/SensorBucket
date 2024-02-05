package tenantsinfra

import (
	"context"
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
	Created time.Time `pagination:"created,DESC"`
}

func (ts *TenantsStore) GetTenantById(id int64) (tenants.Tenant, error) {
	tenant := tenants.Tenant{}
	q := sq.Select(
		"id, name, address, zip_code, city, chamber_of_commerce_id, headquarter_id, archive_time, state, logo, parent_tenant_id").
		From("tenants").Where(sq.Eq{"id": id})
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(ts.db).Query()
	if err != nil {
		return tenants.Tenant{}, err
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
			return tenants.Tenant{}, err
		}
	} else {
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

func (ts *TenantsStore) Update(tenant tenants.Tenant) error {
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
