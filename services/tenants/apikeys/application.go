package apikeys

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type Service struct {
	tenantStore tenantStore
	apiKeyStore apiKeyStore
}

var (
	ErrTenantIsNotValid = fmt.Errorf("tenant is not valid")
	ErrKeyNotFound      = fmt.Errorf("couldnt find key")
)

func (s *Service) GenerateNewApiKey(tenantId int64, expiry *time.Time) (ApiKey, error) {
	tenant, err := s.tenantStore.GetTenantById(tenantId)
	if err != nil {
		return ApiKey{}, err
	}
	if tenant.State != Active {
		return ApiKey{}, ErrTenantIsNotValid
	}
	newApiKey := newApiKey()
	hashed, err := newApiKey.hash()
	if err != nil {
		return ApiKey{}, err
	}
	err = s.apiKeyStore.AddApiKey(tenant.ID, newApiKey.ID, hashed.Value)
	if err != nil {
		return ApiKey{}, err
	}
	return newApiKey, nil
}

func (s *Service) ValidateApiKey(base64IdAndKeyCombination string) (bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(base64IdAndKeyCombination)
	if err != nil {
		// TODO: wrap error in generic invalid base64 error
		return false, err
	}

	// Expected format for the decoded base64 string is 'id:apiKey'
	combo := strings.Split(string(decoded), ":")
	if len(combo) != 2 {
		return false, fmt.Errorf("api key combination format must adhere to 'id:apiKey'")
	}
	apiKeyId, err := strconv.ParseInt(combo[0], 10, 32)
	if err != nil {
		return false, err
	}
	apiKey := combo[1]
	if apiKey == "" {
		return false, fmt.Errorf("api key cannot be empty")
	}
	hashed, err := s.apiKeyStore.GetHashedApiKeyById(apiKeyId)
	if err != nil {
		return false, err
	}
	return hashed.compare(apiKey), nil
}

type apiKeyStore interface {
	AddApiKey(tenantID int64, id int64, hashedKey string) error
	GetHashedApiKeyById(id int64) (HashedApiKey, error)
}

type tenantStore interface {
	GetTenantById(id int64) (Tenant, error)
}
