package tenantsinfra

import (
	"database/sql"
	"errors"
	"fmt"

	"sensorbucket.nl/sensorbucket/services/tenants/sessions"
)

var _ sessions.UserPreferenceStore = (*PSQLTenantStore)(nil)

func (store *PSQLTenantStore) ActiveTenantID(userID string) (int64, error) {
	var id sql.NullInt64
	if err := store.db.Get(&id,
		"SELECT prefered_tenant FROM user_session_preferences WHERE user_id = $1",
		userID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, sessions.ErrPreferenceNotSet
		}
		return 0, fmt.Errorf("in ActiveTenantID PSQL Store, error executing query: %w", err)
	}
	if !id.Valid {
		return 0, sessions.ErrPreferenceNotSet
	}
	return id.Int64, nil
}

func (store *PSQLTenantStore) SetActiveTenantID(userID string, tenantID int64) error {
	var tenantIDValue sql.NullInt64
	if tenantID > 0 {
		tenantIDValue.Int64 = tenantID
		tenantIDValue.Valid = true
	}
	_, err := store.db.Exec(
		`
        INSERT INTO user_session_preferences (user_id, prefered_tenant)
        VALUES ($1, $2)
        ON CONFLICT (user_id)
        DO UPDATE SET
            prefered_tenant = EXCLUDED.prefered_tenant;
        `, userID, tenantIDValue,
	)
	if err != nil {
		return fmt.Errorf("in SetActiveTenant PSQL Store, error executing upsert query: %w", err)
	}
	return nil
}
