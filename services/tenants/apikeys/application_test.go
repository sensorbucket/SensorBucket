package apikeys

import (
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRevokeApiKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		DeleteApiKeyFunc: func(id int64) error {
			assert.Equal(t, int64(665213432), id)
			return nil
		},
	}
	s := &Service{
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
			assert.Equal(t, int64(4343241), id)
			return fmt.Errorf("database error!")
		},
	}
	s := &Service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	err := s.RevokeApiKey(4343241)

	// Assert
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.DeleteApiKeyCalls(), 1)
}

func TestValidateApiKey(t *testing.T) {
	type scene struct {
		Value    string
		Expected bool
		Error    bool
	}

	// Arrange
	scenarios := map[string]scene{
		"invalid base64 string": {
			Value:    "invalid base64 blabla",
			Expected: false,
			Error:    true,
		},
		"invalid decoded format": {
			Value:    asBase64("asdasdjahsdlkoahsd"),
			Expected: false,
			Error:    true,
		},
		"empty api key": {
			Value:    asBase64("1231234:"),
			Expected: false,
			Error:    true,
		},
		"api key id invalid int": {
			Value:    asBase64(("123sad213213:asdasidhlas")),
			Expected: false,
			Error:    true,
		},
		"api key id empty": {
			Value:    asBase64(":asdashdlhasd"),
			Expected: false,
			Error:    true,
		},
	}
	for scenario, input := range scenarios {
		t.Run(scenario, func(t *testing.T) {
			// Act
			s := &Service{}
			res, err := s.ValidateApiKey(input.Value)

			// Assert
			assert.Equal(t, input.Expected, res)
			if input.Error {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateApiKeyErrorOccursWhileRetrievingKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64) (HashedApiKey, error) {
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{}, fmt.Errorf("database error!!")
		},
	}
	s := &Service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.ValidateApiKey(asBase64("43214:somevalidapikey"))

	// Assert
	assert.False(t, res)
	assert.Error(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyInvalidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64) (HashedApiKey, error) {
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{
				Key: Key{
					ID: id,
				},
				Value: "hash is not equal to input key!!",
			}, nil
		},
	}
	s := &Service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.ValidateApiKey(asBase64("43214:someinvalidapikey"))

	// Assert
	assert.False(t, res)
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func TestValidateApiKeyValidKey(t *testing.T) {
	// Arrange
	apiKeyStore := &apiKeyStoreMock{
		GetHashedApiKeyByIdFunc: func(id int64) (HashedApiKey, error) {
			assert.Equal(t, int64(43214), id)
			return HashedApiKey{
				Key: Key{
					ID: id,
				},
				Value: "$2a$10$b1rIBcIIN0SgBjqIIgZp9uPFHbJ0zAcJL27Wu8/kLMlIa0KMXjLua",
			}, nil
		},
	}
	s := &Service{
		apiKeyStore: apiKeyStore,
	}

	// Act
	res, err := s.ValidateApiKey(asBase64("43214:kayJhmgiCNNQAKwtvewxN6BWSTiEINOy"))

	// Assert
	assert.True(t, res)
	assert.NoError(t, err)
	assert.Len(t, apiKeyStore.GetHashedApiKeyByIdCalls(), 1)
}

func asBase64(val string) string {
	return base64.StdEncoding.EncodeToString([]byte(val))
}
