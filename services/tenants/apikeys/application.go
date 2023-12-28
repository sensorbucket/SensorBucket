package apikeys

import (
	"encoding/base64"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
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

func (s *service) ListAPIKeys(filter Filter, p pagination.Request) (*pagination.Page[ApiKeyDTO], error) {
	return s.apiKeyStore.List(filter, p)
}

// Revokes an API key, returns ErrKeyNotFound if the given key was not found in the apiKeyStore
func (s *service) RevokeApiKey(apiKeyId int64) error {
	return s.apiKeyStore.DeleteApiKey(apiKeyId)
}

// Creates a new API key for the given tenant and with the given expiration date.
// Returns the api key as: 'apiKeyId:apiKey' encoded to a base64 string.
// Fails if the tenant is not active
func (s *service) GenerateNewApiKey(name string, tenantId int64, expirationDate *time.Time) (string, error) {
	tenant, err := s.tenantStore.GetTenantById(tenantId)
	if err != nil {
		return "", err
	}
	if tenant.State != Active {
		return "", ErrTenantIsNotValid
	}
	newApiKey, err := newApiKey(name, expirationDate)
	if err != nil {
		return "", err
	}
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
			[]byte(fmt.Sprintf("%d:%s", newApiKey.ID, newApiKey.Secret)))
	return apiKey, nil
}

// Authenticates a given API key. Input must be 'apiKeyId:apiKey' encoded to a base64 string
// API key is valid if it is the correct api key id and api key combination and if the attached tenant is active
func (s *service) AuthenticateApiKey(base64IdAndKeyCombination string) (ApiKeyAuthenticationDTO, error) {
	apiKeyId, apiKey, err := apiKeyAndIdFromBase64(base64IdAndKeyCombination)
	if err != nil {
		return ApiKeyAuthenticationDTO{}, ErrInvalidEncoding
	}
	hashed, err := s.apiKeyStore.GetHashedApiKeyById(apiKeyId, []TenantState{Active})
	if err != nil {
		return ApiKeyAuthenticationDTO{}, err
	}
	if hashed.IsExpired() {
		log.Println("[Info] detected expired API key, deleting")
		if err := s.RevokeApiKey(apiKeyId); err != nil {
			log.Printf("[Warning] couldn't cleanup expired API key: '%s'\n", err)
		}
		return ApiKeyAuthenticationDTO{}, ErrKeyNotFound
	}
	isValid := hashed.compare(apiKey)
	if isValid {
		return ApiKeyAuthenticationDTO{
			TenantID: hashed.TenantID,
		}, nil
	}
	return ApiKeyAuthenticationDTO{}, ErrKeyNotFound
}

type Filter struct {
	TenantID []int64 `schema:"tenant_id"`
}

type ApiKeyAuthenticationDTO struct {
	TenantID   int64 `json:"sub"` // Sub is how Ory Oathkeeper identifies the important information in the response
	Expiration time.Duration
	// TODO: check, how to make ory oathkeeper only use the first value and not 'refresh' this expiration each time a request is authenticated?
	// TODO: is there a diference in flow: just check existing token or  get new token?
}

type ApiKeyDTO struct {
	Name           string     `json:"name"`
	TenantID       int64      `json:"tenant_id"`
	ExpirationDate *time.Time `json:"expiration_date"`
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
	DeleteApiKey(id int64) error
	GetHashedApiKeyById(id int64, stateFilter []TenantState) (HashedApiKey, error)
	List(Filter, pagination.Request) (*pagination.Page[ApiKeyDTO], error)
}

type tenantStore interface {
	GetTenantById(id int64) (Tenant, error)
}
