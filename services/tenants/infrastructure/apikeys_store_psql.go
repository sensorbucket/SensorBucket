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
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var _ apikeys.ApiKeyStore = (*ApiKeyStore)(nil)

func NewAPIKeyStorePSQL(db *sqlx.DB) *ApiKeyStore {
	return &ApiKeyStore{
		db: db,
	}
}

type apiKeyQueryPage struct {
	Created time.Time `pagination:"keys.created,DESC"`
	KeyID   int64     `pagination:"keys.id,DESC"`
}

func (as *ApiKeyStore) List(filter apikeys.Filter, r pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error) {
	var err error
	// Pagination
	cursor, err := pagination.GetCursor[apiKeyQueryPage](r)
	if err != nil {
		return nil, fmt.Errorf("could not getcursor from pagination request: %w", err)
	}

	q := sq.
		Select(
			"keys.id",
			"keys.name",
			"keys.expiration_date",
			"keys.created",
			"keys.tenant_id",
			"tenants.name",
			"keys.permissions",
		).
		From("api_keys keys").
		LeftJoin("tenants on keys.tenant_id = tenants.id")
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"keys.tenant_id": filter.TenantID})
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, fmt.Errorf("could not apply pagination: %w", err)
	}

	list := make([]apikeys.ApiKeyDTO, 0, cursor.Limit)
	// TODO: This is a hack, it grabs the underlying PGX connection to scan the row
	// it is sort of required to scan the array of permission as that column is a postgres
	// special type
	c, err := as.db.Conn(context.Background())
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
		sql, args, err := q.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return err
		}
		rows, err := conn.Query(context.TODO(), sql, args...)
		if err != nil {
			return err
		}
		defer rows.Close()
		for rows.Next() {
			var key apikeys.ApiKeyDTO
			var permissions pgt.FlatArray[auth.Permission]
			err := rows.Scan(
				&key.ID, &key.Name, &key.ExpirationDate, &key.Created, &key.TenantID, &key.TenantName,
				&permissions,
				&cursor.Columns.Created,
				&cursor.Columns.KeyID,
			)
			if err != nil {
				return err
			}
			key.Permissions = auth.Permissions(permissions)
			if err := key.Permissions.Validate(); err != nil {
				return err
			}
			list = append(list, key)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (as *ApiKeyStore) AddApiKey(tenantID int64, permissions auth.Permissions, hashedKey apikeys.HashedApiKey) error {
	// Create the insert API key query
	q := sq.Insert("api_keys").
		Columns("id", "name", "created", "tenant_id", "value", "expiration_date", "permissions").
		Values(
			hashedKey.ID,
			hashedKey.Name,
			time.Now().UTC(),
			tenantID,
			hashedKey.SecretHash,
			hashedKey.ExpirationDate,
			permissions,
		)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	if err != nil {
		return err
	}
	return err
}

// Deletes an API key if found. If the key is not found, ErrKeyNotFound is returned
func (as *ApiKeyStore) DeleteApiKey(id int64) error {
	q := sq.Delete("").From("api_keys").Where("id=?", id)
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	if err != nil {
		return err
	}
	deleted, err := rows.RowsAffected()
	if err != nil {
		return err
	}
	if deleted != 1 {
		return apikeys.ErrKeyNotFound
	}
	return nil
}

// Retrieves the hashed value of an API key, if the key is not found an ErrKeyNotFound is returned.
// Only returns the API key if the given tenant confirms to any state passed in the stateFilter
func (as *ApiKeyStore) GetHashedApiKeyById(id int64, stateFilter []tenants.State) (apikeys.HashedApiKey, error) {
	return as.getAPIKey(func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Where(sq.Eq{
			"keys.id":       id,
			"tenants.state": stateFilter,
		})
	})
}

// Retrieves the hashed value of an API key, if the key is not found an ErrKeyNotFound is returned.
func (as *ApiKeyStore) GetHashedAPIKeyByNameAndTenantID(name string, tenantID int64) (apikeys.HashedApiKey, error) {
	return as.getAPIKey(func(q sq.SelectBuilder) sq.SelectBuilder {
		return q.Where(sq.Eq{"keys.name": name, "keys.tenant_id": tenantID})
	})
}

func (as *ApiKeyStore) getAPIKey(mod func(q sq.SelectBuilder) sq.SelectBuilder) (apikeys.HashedApiKey, error) {
	var key apikeys.HashedApiKey
	var permissions pgt.FlatArray[auth.Permission]
	q := sq.Select(
		"keys.id", "keys.value", "keys.expiration_date", "keys.tenant_id", "keys.permissions",
	).From("api_keys keys").LeftJoin("tenants on keys.tenant_id = tenants.id")
	q = mod(q)
	// TODO: This is a hack, it grabs the underlying PGX connection to scan the row
	// it is sort of required to scan the array of permission as that column is a postgres
	// special type
	c, err := as.db.Conn(context.Background())
	if err != nil {
		return key, fmt.Errorf("in GetTenantMember, could not get raw db conn: %w", err)
	}
	defer c.Close()
	err = c.Raw(func(driverConn any) error {
		stdlibConn, ok := driverConn.(*stdlib.Conn)
		if !ok {
			return errors.New("in GetHashedAPIKeyByNameAndTenantID, expected driverConnection to be of type stdlib.Conn")
		}
		conn := stdlibConn.Conn()
		query, args, err := q.PlaceholderFormat(sq.Dollar).ToSql()
		if err != nil {
			return fmt.Errorf("in GetHashedAPIKeyByNameAndTenantID, could not build query: %w", err)
		}
		row := conn.QueryRow(context.TODO(), query, args...)
		err = row.Scan(
			&key.ID, &key.SecretHash, &key.ExpirationDate, &key.TenantID,
			&permissions,
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return apikeys.ErrKeyNotFound
		}
		if err != nil {
			return fmt.Errorf("in GetHashedAPIKeyByNameAndTenantID, could not scan row: %w", err)
		}
		key.Permissions = auth.Permissions(permissions)
		if err := key.Permissions.Validate(); err != nil {
			return fmt.Errorf("in GetHashedAPIKeyByNameAndTenantID, invalid permissions: %w", err)
		}
		return nil
	})
	if err != nil {
		return key, err
	}
	return key, nil
}

type ApiKeyStore struct {
	db *sqlx.DB
}
