package apikeys

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	ErrTenantIsNotValid = fmt.Errorf("tenant is not valid")
	ErrKeyNotFound      = fmt.Errorf("couldnt find key")
	ErrInvalidEncoding  = fmt.Errorf("API key was sent using an invalid encoding")
)

func NewAPIKeyService(tenantStore tenantStore, apiKeyStore apiKeyStore) *service {
	return &service{
		tenantStore: tenantStore,
		apiKeyStore: apiKeyStore,
	}
}

// Revokes an API key, returns ErrKeyNotFound if the given key was not found in the apiKeyStore
func (s *service) RevokeApiKey(base64IdAndKeyCombination string) error {
	apiKeyId, _, err := apiKeyAndIdFromBase64(base64IdAndKeyCombination)
	if err != nil {
		return ErrInvalidEncoding
	}

	res, err := s.apiKeyStore.DeleteApiKey(apiKeyId)
	if err != nil {
		return err
	}
	if !res {
		return ErrKeyNotFound
	}
	return nil
}

// Creates a new API key for the given tenant and with the given expiration date.
// Returns the api key as: 'apiKeyId:apiKey' encoded to a base64 string.
// Fails if the tenant is not active
func (s *service) GenerateNewApiKey(tenantId int64, expirationDate *time.Time) (string, error) {
	tenant, err := s.tenantStore.GetTenantById(tenantId)
	if err != nil {
		return "", err
	}
	if tenant.State != Active {
		return "", ErrTenantIsNotValid
	}
	newApiKey := newApiKey(expirationDate)
	hashed, err := newApiKey.hash()
	if err != nil {
		return "", err
	}
	err = s.apiKeyStore.AddApiKey(tenant.ID, hashed)
	if err != nil {
		return "", err
	}
	apiKey :=
		base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(
			[]byte(fmt.Sprintf("%d:%s", newApiKey.ID, newApiKey.Value)))
	return apiKey, nil
}

// Validates a given API key. Input must be 'apiKeyId:apiKey' encoded to a base64 string
// API key is valid if it is the correct api key id and api key combination and if the attached tenant is active
func (s *service) ValidateApiKey(base64IdAndKeyCombination string) (bool, error) {
	apiKeyId, apiKey, err := apiKeyAndIdFromBase64(base64IdAndKeyCombination)
	if err != nil {
		return false, ErrInvalidEncoding
	}
	hashed, err := s.apiKeyStore.GetHashedApiKeyById(apiKeyId)
	if err != nil {
		return false, err
	}
	if hashed.IsExpired() {
		log.Println("[Info] detected expired API key, deleting")
		if err := s.RevokeApiKey(base64IdAndKeyCombination); err != nil {
			log.Printf("[Warning] couldn't cleanup expired API key: '%s'\n", err)
		}
		return false, nil
	}
	return hashed.compare(apiKey), nil
}

func apiKeyAndIdFromBase64(base64Src string) (int64, string, error) {
	decoded, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(base64Src)
	if err != nil {
		return -1, "", err
	}
	// Expected format for the decoded base64 string is 'id:apiKey'
	combo := strings.Split(string(decoded), ":")
	if len(combo) != 2 {
		return -1, "", fmt.Errorf("api key combination format must adhere to 'id:apiKey'")
	}
	apiKeyId, err := strconv.ParseInt(combo[0], 10, 64)
	if err != nil {
		return -1, "", err
	}
	apiKey := combo[1]
	if apiKey == "" {
		return -1, "", fmt.Errorf("api key cannot be empty")
	}
	return apiKeyId, apiKey, nil
}

type service struct {
	tenantStore tenantStore
	apiKeyStore apiKeyStore
}

type apiKeyStore interface {
	AddApiKey(tenantID int64, hashedApiKey HashedApiKey) error
	DeleteApiKey(id int64) (bool, error)
	GetHashedApiKeyById(id int64) (HashedApiKey, error)
}

type tenantStore interface {
	GetTenantById(id int64) (Tenant, error)
}
