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
		Select("keys.id", "keys.name", "keys.expiration_date", "keys.created_at", "keys.tenant_id", "tenant.name", "permissions.permission").
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

	// TODO: limit wont work with the way permissois are joined now

	// For each key there are multiple records so all the permissions can be listed
	// Track at which API key the rows are currently so we can add the correct permissions and move to the next key
	list := make([]apikeys.ApiKeyDTO, 0, cursor.Limit)
	currentId := -1
	lastId := -1
	currentPermissions := []string{}
	for rows.Next() {
		key := apikeys.ApiKeyDTO{}
		permission := ""
		err = rows.Scan(
			&key.ID,
			&key.Name,
			&key.ExpirationDate,
			&key.Created,
			&key.TenantID,
			&key.TenantName,
			&permission,
			&cursor.Columns.Tenant,
			&cursor.Columns.Created,
		)
		if err != nil {
			return nil, err
		}
		if lastId != currentId {
			// Started scanning new API key record
			currentPermissions = []string{}
		}
		currentPermissions = append(currentPermissions, permission)
		lastId = currentId
		if len(list) > 0 && list[len(list)-1].ID == key.ID {
			// Still at the same key, append the permission to the last key
			list[len(list)-1].Permissions = append(list[len(list)-1].Permissions, permission)
		} else {
			// Otherwise the result set arrived at a new api key
			key.Permissions = []string{permission}
			list = append(list, key)
		}
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
	// TODO: how secure is this query?
	q := sq.
		Select("api_keys.id, api_keys.value, api_keys.tenant_id, api_keys.expiration_date", "api_key_permissions.permission").
		From("api_keys").
		Where(sq.Eq{"api_keys.id": id}).
		Join("api_key_permissions on api_keys.id = api_key_permissions.api_key_id")
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
	for rows.Next() {
		permission := ""
		err = rows.Scan(
			&k.ID,
			&k.SecretHash,
			&k.TenantID,
			&k.ExpirationDate,
			&permission,
		)
		if err != nil {
			return apikeys.HashedApiKey{}, err
		}
		k.Permissions = append(k.Permissions, permission)
	}
	if k.ID == 0 {
		return apikeys.HashedApiKey{}, apikeys.ErrKeyNotFound
	}
	return k, nil
}

type apiKeyStore struct {
	db *sqlx.DB
}
