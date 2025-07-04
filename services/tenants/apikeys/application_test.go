package apikeys_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"sensorbucket.nl/sensorbucket/pkg/auth"
	"sensorbucket.nl/sensorbucket/services/tenants/apikeys"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

func TestGenerateNewApiKeyCreatesNewApiKey(t *testing.T) {
	// Arrange
	exp := time.Date(2024, 12, 9, 33, 12, 50, 300, time.UTC)
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(905), id)
			return &tenants.Tenant{
				ID:    905,
				State: tenants.Active,
			}, nil
		},
	}
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedAPIKeyByNameAndTenantIDFunc: func(ctx context.Context, _ string, tenantID int64) (apikeys.HashedApiKey, error) {
			assert.Equal(t, int64(905), tenantID)
			return apikeys.HashedApiKey{}, apikeys.ErrKeyNotFound
		},
		AddApiKeyFunc: func(ctx context.Context, tenantID int64, permissions auth.Permissions, hashedApiKey apikeys.HashedApiKey) error {
			assert.Equal(t, int64(905), tenantID)
			assert.Equal(t, auth.Permissions{auth.READ_DEVICES}, permissions)
			assert.NotNil(t, hashedApiKey.ExpirationDate)
			assert.Equal(t, exp, *hashedApiKey.ExpirationDate)
			assert.NotEmpty(t, hashedApiKey.SecretHash)
			assert.NotEqual(t, 0, hashedApiKey.ID)
			return nil
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)
	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{auth.READ_DEVICES},
		&exp,
	)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, apiKeyStore.GetHashedAPIKeyByNameAndTenantIDCalls(), 1)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
	assert.Len(t, apiKeyStore.AddApiKeyCalls(), 1)
}

func TestGenerateNewAPIKeyNameAndTenantCombinationNotUnique(t *testing.T) {
	// Arrange
	exp := time.Date(2024, 12, 9, 33, 12, 50, 300, time.UTC)
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(905), id)
			return &tenants.Tenant{
				ID:    905,
				State: tenants.Active,
			}, nil
		},
	}
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedAPIKeyByNameAndTenantIDFunc: func(ctx context.Context, name string, tenantID int64) (apikeys.HashedApiKey, error) {
			assert.Equal(t, int64(905), tenantID)
			return apikeys.HashedApiKey{
				Key: apikeys.Key{
					ID:   2431,
					Name: "already exists!",
				},
			}, nil
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{auth.READ_API_KEYS},
		&exp,
	)

	// Assert
	assert.ErrorIs(t, err, apikeys.ErrKeyNameTenantIDCombinationNotUnique)
	assert.Empty(t, res)
	assert.Len(t, apiKeyStore.GetHashedAPIKeyByNameAndTenantIDCalls(), 1)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
}

func TestGenerateNewAPIKeyCheckCombinationUniqueErrorOccurs(t *testing.T) {
	// Arrange
	exp := time.Date(2024, 12, 9, 33, 12, 50, 300, time.UTC)
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(905), id)
			return &tenants.Tenant{
				ID:    905,
				State: tenants.Active,
			}, nil
		},
	}
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedAPIKeyByNameAndTenantIDFunc: func(ctx context.Context, name string, tenantID int64) (apikeys.HashedApiKey, error) {
			assert.Equal(t, int64(905), tenantID)
			return apikeys.HashedApiKey{}, fmt.Errorf("weird db error!")
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{auth.READ_DEVICES},
		&exp,
	)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, res)
	assert.Len(t, apiKeyStore.GetHashedAPIKeyByNameAndTenantIDCalls(), 1)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
}

func TestGenerateNewApiKeyErrorOccursWhileAddingApiKeyToStore(t *testing.T) {
	// Arrange
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(905), id)
			return &tenants.Tenant{
				ID:    905,
				State: tenants.Active,
			}, nil
		},
	}
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedAPIKeyByNameAndTenantIDFunc: func(ctx context.Context, name string, tenantID int64) (apikeys.HashedApiKey, error) {
			assert.Equal(t, int64(905), tenantID)
			return apikeys.HashedApiKey{}, apikeys.ErrKeyNotFound
		},
		AddApiKeyFunc: func(ctx context.Context, tenantID int64, permissions auth.Permissions, hashedApiKey apikeys.HashedApiKey) error {
			assert.Equal(t, int64(905), tenantID)
			assert.Equal(t, auth.Permissions{auth.READ_DEVICES}, permissions)
			assert.NotEmpty(t, hashedApiKey.SecretHash)
			assert.NotEqual(t, 0, hashedApiKey.ID)
			return fmt.Errorf("weird database error!!")
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{auth.READ_DEVICES},
		nil,
	)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
	assert.Len(t, apiKeyStore.AddApiKeyCalls(), 1)
}

func TestGenerateNewApiKeyPermissionsContains1InvalidPermission(t *testing.T) {
	// Arrange
	tenantStore := &TenantStoreMock{}
	apiKeyStore := &ApiKeyStoreMock{}
	s := apikeys.NewAPIKeyService(tenantStore, apiKeyStore)

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{
			auth.READ_API_KEYS,
			auth.READ_DEVICES,
			auth.Permission("invalidpermission"),
		},
		nil,
	)

	// Assert
	assert.ErrorIs(t, err, apikeys.ErrPermissionsInvalid)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 0)
	assert.Len(t, apiKeyStore.AddApiKeyCalls(), 0)
}

func TestGenerateNewApiKeyErrorOccursWhenRetrievingTenant(t *testing.T) {
	// Arrange
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(905), id)
			return &tenants.Tenant{
				State: tenants.Active,
			}, fmt.Errorf("weird database error!")
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, &ApiKeyStoreMock{})

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		905,
		auth.Permissions{auth.READ_DEVICES},
		nil,
	)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
}

func TestGenerateNewApiKeyTenantDoesNotExist(t *testing.T) {
	// Arrange
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(334), id)
			return &tenants.Tenant{}, apikeys.ErrTenantIsNotValid
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, &ApiKeyStoreMock{})

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		334,
		auth.Permissions{auth.READ_DEVICES},
		nil,
	)

	// Assert
	assert.ErrorIs(t, err, apikeys.ErrTenantIsNotValid)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
}

func TestGenerateNewApiKeyTenantIsNottenantsActive(t *testing.T) {
	// Arrange
	tenantStore := &TenantStoreMock{
		GetTenantByIDFunc: func(ctx context.Context, id int64) (*tenants.Tenant, error) {
			assert.Equal(t, int64(334), id)
			return &tenants.Tenant{
				State: tenants.Archived,
			}, nil
		},
	}
	s := apikeys.NewAPIKeyService(tenantStore, &ApiKeyStoreMock{})

	// Act
	res, err := s.GenerateNewApiKey(
		context.Background(),
		"whatever",
		334,
		auth.Permissions{auth.READ_DEVICES},
		nil,
	)

	// Assert
	assert.ErrorIs(t, err, apikeys.ErrTenantIsNotValid)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIDCalls(), 1)
}

func TestRevokeApiKeyDeletesKey(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		DeleteApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(665213432), id)
			return nil
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	err := s.RevokeApiKey(context.Background(), 665213432)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestRevokeApiKeyErrorOccurs(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		DeleteApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(83245345), id)
			return fmt.Errorf("weird error!!")
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	err := s.RevokeApiKey(context.Background(), 83245345)

	// Assert
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestRevokeApiKeyWasNotDeletedByStore(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		DeleteApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(83245345), id)
			return apikeys.ErrKeyNotFound
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	err := s.RevokeApiKey(context.Background(), 83245345)

	// Assert
	assert.ErrorIs(t, err, apikeys.ErrKeyNotFound)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyInvalidEncoding(t *testing.T) {
	// Arrange
	scenarios := map[string]string{
		"invalid base64 string":  "invalid base64 blabla",
		"invalid decoded format": asBase64("asdasdjahsdlkoahsd"),
		"empty api key":          asBase64("1231234:"),
		"api key id invalid int": asBase64("123sad213213:asdasidhlas"),
		"api key id empty":       asBase64(":asdashdlhasd"),
	}
	for scenario, input := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			// Act
			s := &apikeys.Service{}
			res, err := s.AuthenticateApiKey(context.Background(), input)

			// Assert
			assert.EqualValues(t, 0, res.TenantID)
			assert.ErrorIs(t, err, apikeys.ErrInvalidEncoding)
		})
	}
}

func TestValidateApiKeyErrorOccursWhileRetrievingKey(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(ctx context.Context, id int64, stateFilter apikeys.APIKeyFilter) (apikeys.HashedApiKey, error) {
			assert.Equal(t, []tenants.State{tenants.Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return apikeys.HashedApiKey{}, fmt.Errorf("database error!!")
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	res, err := s.AuthenticateApiKey(context.Background(), asBase64("43214:somevalidapikey"))

	// Assert
	assert.EqualValues(t, 0, res.TenantID)
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyInvalidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(ctx context.Context, id int64, stateFilter apikeys.APIKeyFilter) (apikeys.HashedApiKey, error) {
			assert.Equal(t, []tenants.State{tenants.Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return apikeys.HashedApiKey{
				Key: apikeys.Key{
					ID: id,
				},
				TenantID:   823,
				SecretHash: "hash is not equal to input key!!",
			}, nil
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	res, err := s.AuthenticateApiKey(context.Background(), asBase64("43214:someinvalidapikey"))

	// Assert
	assert.EqualValues(t, 0, res.TenantID)
	assert.ErrorIs(t, err, apikeys.ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyKeyIsExpired(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(ctx context.Context, id int64, stateFilter apikeys.APIKeyFilter) (apikeys.HashedApiKey, error) {
			assert.Equal(t, []tenants.State{tenants.Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			t := time.Date(2010, 11, 12, 8, 37, 3, 500, time.Local)
			return apikeys.HashedApiKey{
				Key: apikeys.Key{
					ID:             id,
					ExpirationDate: &t,
				},
				TenantID:   678,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
		DeleteApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(43214), id)
			return nil
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	res, err := s.AuthenticateApiKey(
		context.Background(),
		asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"),
	)

	// Assert
	assert.EqualValues(t, 0, res.TenantID)
	assert.ErrorIs(t, err, apikeys.ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyKeyIsExpiredDeleteErrorOccurs(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(ctx context.Context, id int64, stateFilter apikeys.APIKeyFilter) (apikeys.HashedApiKey, error) {
			assert.Equal(t, []tenants.State{tenants.Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			t := time.Date(2010, 11, 12, 8, 37, 3, 500, time.Local)
			return apikeys.HashedApiKey{
				Key: apikeys.Key{
					ID:             id,
					ExpirationDate: &t,
				},
				TenantID:   123,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
		DeleteApiKeyFunc: func(ctx context.Context, id int64) error {
			assert.Equal(t, int64(43214), id)
			return fmt.Errorf("weird database error")
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	res, err := s.AuthenticateApiKey(
		context.Background(),
		asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"),
	)

	// Assert
	assert.EqualValues(t, 0, res.TenantID)
	assert.ErrorIs(t, err, apikeys.ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyValidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &ApiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(ctx context.Context, id int64, stateFilter apikeys.APIKeyFilter) (apikeys.HashedApiKey, error) {
			assert.Equal(t, []tenants.State{tenants.Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return apikeys.HashedApiKey{
				Key: apikeys.Key{
					ID: id,
				},
				TenantID:   534,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
	}
	s := apikeys.NewAPIKeyService(&TenantStoreMock{}, apiKeyStore)

	// Act
	res, err := s.AuthenticateApiKey(
		context.Background(),
		asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"),
	)

	// Assert
	assert.EqualValues(t, 534, res.TenantID)
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func asBase64(val string) string {
	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(val))
}
