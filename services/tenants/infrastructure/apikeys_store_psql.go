package tenantsinfra

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
)

func NewAPIKeyStorePSQL(db *sqlx.DB) *apiKeyStore {
	return &apiKeyStore{
		db: db,
	}
}

type apiKeyQueryPage struct {
	Created time.Time `pagination:"created,DESC"`
}

func (as *apiKeyStore) List(filter apikeys.Filter, r pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error) {
	var err error

	// Pagination
	cursor, err := pagination.GetCursor[apiKeyQueryPage](r)
	if err != nil {
		return nil, err
	}

	q := sq.Select("name, expiration_date, tenant_id").Distinct().From("api_keys")
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"tenant_id": filter.TenantID})
	}
	q, err = pagination.Apply(q, cursor)
	if err != nil {
		return nil, err
	}
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]apikeys.ApiKeyDTO, 0, cursor.Limit)
	for rows.Next() {
		key := apikeys.ApiKeyDTO{}
		err = rows.Scan(
			&key.Name,
			&key.ExpirationDate,
			&key.TenantID,
			&cursor.Columns.Created,
		)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		list = append(list, key)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (as *apiKeyStore) AddApiKey(tenantID int64, hashedKey apikeys.HashedApiKey) error {
	q := sq.Insert("api_keys").
		Columns("id", "name", "created", "tenant_id", "value", "expiration_date").
		Values(
			hashedKey.ID,
			hashedKey.Name,
			time.Now().UTC(),
			tenantID,
			hashedKey.Value,
			hashedKey.ExpirationDate)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	return err
}

func (as *apiKeyStore) DeleteApiKey(id int64) (bool, error) {
	q := sq.Delete("").From("api_keys").Where("id=?", id)
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	if err != nil {
		return false, err
	}
	deleted, err := rows.RowsAffected()
	if err != nil {
		return false, err
	}
	return deleted == 1, err
}

func (as *apiKeyStore) GetHashedApiKeyById(id int64) (apikeys.HashedApiKey, error) {
	q := sq.Select("id, value, expiration_date").From("api_keys").Where("id=?", id)
	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Query()
	if err != nil {
		return apikeys.HashedApiKey{}, err
	}
	k := apikeys.HashedApiKey{}
	for rows.Next() {
		err = rows.Scan(
			&k.ID,
			&k.Value,
			&k.ExpirationDate)
		if err != nil {
			return apikeys.HashedApiKey{}, err
		}
	}
	return k, nil
}

type apiKeyStore struct {
	db *sqlx.DB
}
