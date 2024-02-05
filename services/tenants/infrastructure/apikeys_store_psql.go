package tenantsinfra

import (
	"time"

	sq "github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
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
			hashedKey.SecretHash,
			hashedKey.ExpirationDate)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	return err
}

// Deletes an API key if found. If the key is not found, ErrKeyNotFound is returned
func (as *apiKeyStore) DeleteApiKey(id int64) error {
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
func (as *apiKeyStore) GetHashedApiKeyById(id int64, stateFilter []tenants.State) (apikeys.HashedApiKey, error) {
	q := sq.
		Select("api_keys.id, api_keys.value, api_keys.tenant_id, api_keys.expiration_date").
		From("api_keys").
		Where(sq.Eq{"api_keys.id": id})
	if len(stateFilter) > 0 {
		q = q.Join("tenants on api_keys.tenant_id = tenants.id").
			Where(sq.Eq{"tenants.state": stateFilter})
	}

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Query()
	if err != nil {
		return apikeys.HashedApiKey{}, err
	}
	defer rows.Close()

	k := apikeys.HashedApiKey{}
	if rows.Next() {
		err = rows.Scan(
			&k.ID,
			&k.SecretHash,
			&k.TenantID,
			&k.ExpirationDate)
		if err != nil {
			return apikeys.HashedApiKey{}, err
		}
	} else {
		return apikeys.HashedApiKey{}, apikeys.ErrKeyNotFound
	}
	return k, nil
}

type apiKeyStore struct {
	db *sqlx.DB
}
