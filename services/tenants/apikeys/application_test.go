package apikeys

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGenerateNewApiKeyCreatesNewApiKey(t *testing.T) {
	// Arrange
	exp := time.Date(2024, 12, 9, 33, 12, 50, 300, time.UTC)
	tenantStore := &tenantStoreMock{
		GetTenantByIdFunc: func(id int64) (Tenant, error) {
			assert.Equal(t, int64(905), id)
			return Tenant{
				ID:    905,
				State: Active,
			}, nil
		},
	}
	apiKeyStore := &apiKeyStoreMock{
		AddApiKeyFunc: func(tenantID int64, hashedApiKey HashedApiKey) error {
			assert.Equal(t, int64(905), tenantID)
			assert.NotNil(t, hashedApiKey.ExpirationDate)
			assert.Equal(t, exp, *hashedApiKey.ExpirationDate)
			assert.NotEmpty(t, hashedApiKey.SecretHash)
			assert.NotEqual(t, 0, hashedApiKey.Key.ID)
			return nil
		},
	}
	s := &service{
		tenantStore: tenantStore,
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.GenerateNewApiKey("whatever", 905, &exp)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, res)
	assert.Len(t, tenantStore.GetTenantByIdCalls(), 1)
	assert.Len(t, apiKeyStore.AddApiKeyCalls(), 1)
}

func TestGenerateNewApiKeyErrorOccursWhileAddingApiKeyToStore(t *testing.T) {
	// Arrange
	tenantStore := &tenantStoreMock{
		GetTenantByIdFunc: func(id int64) (Tenant, error) {
			assert.Equal(t, int64(905), id)
			return Tenant{
				ID:    905,
				State: Active,
			}, nil
		},
	}
	apiKeyStore := &apiKeyStoreMock{
		AddApiKeyFunc: func(tenantID int64, hashedApiKey HashedApiKey) error {
			assert.Equal(t, int64(905), tenantID)
			assert.NotEmpty(t, hashedApiKey.SecretHash)
			assert.NotEqual(t, 0, hashedApiKey.Key.ID)
			return fmt.Errorf("weird database error!!")
		},
	}
	s := &service{
		tenantStore: tenantStore,
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.GenerateNewApiKey("whatever", 905, nil)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIdCalls(), 1)
	assert.Len(t, apiKeyStore.AddApiKeyCalls(), 1)
}

func TestGenerateNewApiKeyErrorOccursWhenRetrievingTenant(t *testing.T) {
	// Arrange
	tenantStore := &tenantStoreMock{
		GetTenantByIdFunc: func(id int64) (Tenant, error) {
			assert.Equal(t, int64(905), id)
			return Tenant{
				State: Active,
			}, fmt.Errorf("weird database error!")
		},
	}
	s := &service{
		tenantStore: tenantStore,
	}

	// Act
	res, err := s.GenerateNewApiKey("whatever", 905, nil)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIdCalls(), 1)
}

func TestGenerateNewApiKeyTenantDoesNotExist(t *testing.T) {
	// Arrange
	tenantStore := &tenantStoreMock{
		GetTenantByIdFunc: func(id int64) (Tenant, error) {
			assert.Equal(t, int64(334), id)
			return Tenant{}, ErrTenantIsNotValid
		},
	}
	s := &service{
		tenantStore: tenantStore,
	}

	// Act
	res, err := s.GenerateNewApiKey("whatever", 334, nil)

	// Assert
	assert.ErrorIs(t, err, ErrTenantIsNotValid)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIdCalls(), 1)
}

func TestGenerateNewApiKeyTenantIsNotActive(t *testing.T) {
	// Arrange
	tenantStore := &tenantStoreMock{
		GetTenantByIdFunc: func(id int64) (Tenant, error) {
			assert.Equal(t, int64(334), id)
			return Tenant{
				State: Archived,
			}, nil
		},
	}
	s := &service{
		tenantStore: tenantStore,
	}

	// Act
	res, err := s.GenerateNewApiKey("whatever", 334, nil)

	// Assert
	assert.ErrorIs(t, err, ErrTenantIsNotValid)
	assert.Empty(t, res)
	assert.Len(t, tenantStore.GetTenantByIdCalls(), 1)
}

func TestRevokeApiKeyDeletesKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(665213432), id)
			return nil
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	err := s.RevokeApiKey(665213432)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestRevokeApiKeyErrorOccurs(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(83245345), id)
			return fmt.Errorf("weird error!!")
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	err := s.RevokeApiKey(83245345)

	// Assert
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestRevokeApiKeyWasNotDeletedByStore(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(83245345), id)
			return ErrKeyNotFound
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	err := s.RevokeApiKey(83245345)

	// Assert
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyInvalidEncoding(t *testing.T) {
	type scene struct {
		Value    string
		Expected bool
		Error    bool
	}

	// Arrange
	scenarios := map[string]string{
		"invalid base64 string":  "invalid base64 blabla",
		"invalid decoded format": asBase64("asdasdjahsdlkoahsd"),
		"empty api key":          asBase64("1231234:"),
		"api key id invalid int": asBase64(("123sad213213:asdasidhlas")),
		"api key id empty":       asBase64(":asdashdlhasd"),
	}
	for scenario, input := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			// Act
			s := &service{}
			res, err := s.AuthenticateApiKey(input)

			// Assert
			assert.Equal(t, "", res.TenantID)
			assert.ErrorIs(t, err, ErrInvalidEncoding)
		})
	}
}

func TestValidateApiKeyErrorOccursWhileRetrievingKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64, stateFilter []TenantState) (HashedApiKey, error) {
			assert.Equal(t, []TenantState{Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{}, fmt.Errorf("database error!!")
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.AuthenticateApiKey(asBase64("43214:somevalidapikey"))

	// Assert
	assert.Equal(t, "", res.TenantID)
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyInvalidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64, stateFilter []TenantState) (HashedApiKey, error) {
			assert.Equal(t, []TenantState{Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{
				Key: Key{
					ID: id,
				},
				TenantID:   823,
				SecretHash: "hash is not equal to input key!!",
			}, nil
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.AuthenticateApiKey(asBase64("43214:someinvalidapikey"))

	// Assert
	assert.Equal(t, "", res.TenantID)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyKeyIsExpired(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64, stateFilter []TenantState) (HashedApiKey, error) {
			assert.Equal(t, []TenantState{Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			t := time.Date(2010, 11, 12, 8, 37, 3, 500, time.Local)
			return HashedApiKey{
				Key: Key{
					ID:             id,
					ExpirationDate: &t,
				},
				TenantID:   678,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(43214), id)
			return nil
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.AuthenticateApiKey(asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"))

	// Assert
	assert.Equal(t, "", res.TenantID)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyKeyIsExpiredDeleteErrorOccurs(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64, stateFilter []TenantState) (HashedApiKey, error) {
			assert.Equal(t, []TenantState{Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			t := time.Date(2010, 11, 12, 8, 37, 3, 500, time.Local)
			return HashedApiKey{
				Key: Key{
					ID:             id,
					ExpirationDate: &t,
				},
				TenantID:   123,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(43214), id)
			return fmt.Errorf("weird database error")
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.AuthenticateApiKey(asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"))

	// Assert
	assert.Equal(t, "", res.TenantID)
	assert.ErrorIs(t, err, ErrKeyNotFound)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKeyValidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64, stateFilter []TenantState) (HashedApiKey, error) {
			assert.Equal(t, []TenantState{Active}, stateFilter)
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{
				Key: Key{
					ID: id,
				},
				TenantID:   534,
				SecretHash: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
	}
	s := &service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.AuthenticateApiKey(asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"))

	// Assert
	assert.Equal(t, "534", res.TenantID)
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func asBase64(val string) string {
	return base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString([]byte(val))
}
