package tenantsinfra

import (
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
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

	keyQuery := sq.Select("*").From("api_keys keys")
	if len(filter.TenantID) > 0 {
		keyQuery = keyQuery.Where(sq.Eq{"keys.tenant_id": filter.TenantID})
	}
	keyQuery, err = pagination.Apply(keyQuery, cursor)
	if err != nil {
		return nil, fmt.Errorf("could not apply pagination: %w", err)
	}

	q := sq.
		Select(
			"keys.id",
			"keys.name",
			"keys.expiration_date",
			"keys.created",
			"keys.tenant_id",
			"tenants.name",
			"permissions.permission",
			"keys.created",
			"keys.id",
		).
		FromSelect(keyQuery, "keys").
		LeftJoin("tenants on keys.tenant_id = tenants.id").
		LeftJoin("api_key_permissions permissions on keys.id = permissions.api_key_id")

	fmt.Printf("sq.DebugSqlizer(q): %v\n", sq.DebugSqlizer(q))

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Query()
	if err != nil {
		return nil, fmt.Errorf("error running database query: %w", err)
	}
	defer rows.Close()

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
			&cursor.Columns.Created,
			&cursor.Columns.KeyID,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row into api key w/ permission: %w", err)
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

func (as *ApiKeyStore) AddApiKey(tenantID int64, permissions auth.Permissions, hashedKey apikeys.HashedApiKey) error {
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
	q := sq.
		Select(
			"key.id", "key.name", "key.value", "key.expiration_date",
			"key.tenant_id",
			"perm.permission",
		).
		Where(sq.Eq{"key.id": id}).
		From("api_keys key").
		LeftJoin("api_key_permissions perm on perm.api_key_id = key.id")
	if len(stateFilter) > 0 {
		q = q.
			LeftJoin("tenants on tenants.id = key.tenant_id").
			Where(sq.Eq{"tenants.state": stateFilter})
	}
	fmt.Printf("sq.DebugSqlizer(q): %v\n", sq.DebugSqlizer(q))

	rows, err := q.PlaceholderFormat(sq.Dollar).RunWith(as.db).Query()
	if err != nil {
		return apikeys.HashedApiKey{}, err
	}
	defer rows.Close()

	key := apikeys.HashedApiKey{
		Permissions: make([]string, 0),
	}
	for rows.Next() {
		var permission string
		err = rows.Scan(
			&key.ID,
			&key.Name,
			&key.SecretHash,
			&key.ExpirationDate,
			&key.TenantID,
			&permission,
		)
		if err != nil {
			return key, err
		}
		key.Permissions = append(key.Permissions, permission)
	}
	if key.ID == 0 {
		return key, apikeys.ErrKeyNotFound
	}
	return key, nil
}

// Retrieves the hashed value of an API key, if the key is not found an ErrKeyNotFound is returned.
func (as *ApiKeyStore) GetHashedAPIKeyByNameAndTenantID(name string, tenantID int64) (apikeys.HashedApiKey, error) {
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

type ApiKeyStore struct {
	db *sqlx.DB
}
