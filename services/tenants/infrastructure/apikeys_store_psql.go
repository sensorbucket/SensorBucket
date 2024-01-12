package tenantsinfra

import (
	"fmt"
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
	Tenant  int64     `pagination:"api_keys.tenant_id,DESC"`
	Created time.Time `pagination:"api_keys.created,DESC"`
}

func (as *apiKeyStore) List(filter apikeys.Filter, r pagination.Request) (*pagination.Page[apikeys.ApiKeyDTO], error) {
	var err error

	// Pagination
	cursor, err := pagination.GetCursor[apiKeyQueryPage](r)
	if err != nil {
		return nil, err
	}

	q := sq.
		Select("keys.id", "keys.name", "keys.expiration_date", "keys.tenant_id", "tenant.name", "permissions.permission").
		From("api_keys keys").
		LeftJoin("permissions on keys.id = permissions.api_key_id").
		RightJoin("tenants on keys.tenant_id = tenants.id")
	if len(filter.TenantID) > 0 {
		q = q.Where(sq.Eq{"api_keys.tenant_id": filter.TenantID})
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
	currentId := -1
	lastId := -1
	currentPermissions := []string{}
	for rows.Next() {
		key := apikeys.ApiKeyDTO{}
		permission := ""
		err = rows.Scan(
<<<<<<< HEAD
			&key.ID,
			&key.Name,
			&key.ExpirationDate,
			&key.TenantID,
			&key.TenantName,
			&cursor.Columns.Tenant,
			&key.Created,
=======
			&currentId,
			&key.Name,
			&key.ExpirationDate,
			&key.TenantID,
			&permission,
			&cursor.Columns.Created,
>>>>>>> 7875bad (update)
		)
		if err != nil {
			return nil, err
		}
<<<<<<< HEAD

		cursor.Columns.Created = key.Created

=======
		if lastId != currentId {
			// Started scanning new API key record
			currentPermissions = []string{}
		}
		currentPermissions = append(currentPermissions, permission)
		lastId = currentId
>>>>>>> 7875bad (update)
		list = append(list, key)
	}
	page := pagination.CreatePageT(list, cursor)
	return &page, nil
}

func (as *apiKeyStore) AddApiKey(tenantID int64, permissions []string, hashedKey apikeys.HashedApiKey) error {

	// Create the insert statement for the permissions which must ran with the insert API key query
	apiKeyPermissionsQ := sq.Insert("api_key_permissions").
		Columns("permission", "api_key_id")
	for _, permission := range permissions {
		apiKeyPermissionsQ = apiKeyPermissionsQ.Values(permission, sq.Select("id").From("new_api_key").Prefix("(").Suffix(")"))
	}

	// Create the insert API key query
	q := sq.Insert("api_keys").
		Columns("id", "name", "created", "tenant_id", "value", "expiration_date").
		Values(
			hashedKey.ID,
			hashedKey.Name,
			time.Now().UTC(),
			tenantID,
			hashedKey.SecretHash,
			hashedKey.ExpirationDate).
		Prefix("WITH new_api_key AS (").
		Suffix("RETURNING \"id\")").

		// Run the API key permission query along with the insert API key query
		SuffixExpr(apiKeyPermissionsQ)
	_, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Exec()
	if err != nil {
		return err
	}
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
	defer rows.Close()
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

// Retrieves the hashed value of an API key, if the key is not found an ErrKeyNotFound is returned.
func (as *apiKeyStore) GetHashedAPIKeyByNameAndTenantID(name string, tenantID int64) (apikeys.HashedApiKey, error) {
	q := sq.
		Select("id, value,  tenant_id,  expiration_date").
		From("api_keys").
		Where(sq.Eq{"name": name, "tenant_id": tenantID})
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
