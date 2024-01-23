package apikeys

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sensorbucket.nl/sensorbucket/internal/pagination"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var (
	ErrTenantIsNotValid                    = fmt.Errorf("tenant is not valid")
	ErrKeyNotFound                         = fmt.Errorf("couldnt find key")
	ErrInvalidEncoding                     = fmt.Errorf("API key was sent using an invalid encoding")
	ErrKeyNameTenantIDCombinationNotUnique = web.NewError(http.StatusBadRequest, "API Key with name already exists for this tenant", "API_KEY_NAME_TENANT_COMBO_NOT_UNIQUE")
)

func NewAPIKeyService(tenantStore tenantStore, apiKeyStore apiKeyStore) *Service {
	return &Service{
		tenantStore: tenantStore,
		apiKeyStore: apiKeyStore,
	}
}

func (s *Service) ListAPIKeys(filter Filter, p pagination.Request) (*pagination.Page[ApiKeyDTO], error) {
	return s.apiKeyStore.List(filter, p)
}

// Revokes an API key, returns ErrKeyNotFound if the given key was not found in the apiKeyStore
func (s *Service) RevokeApiKey(apiKeyId int64) error {
	return s.apiKeyStore.DeleteApiKey(apiKeyId)
}

// Creates a new API key for the given tenant and with the given expiration date.
// Returns the api key as: 'apiKeyId:apiKey' encoded to a base64 string.
// Fails if the tenant is not active
func (s *Service) GenerateNewApiKey(name string, tenantId int64, expirationDate *time.Time) (string, error) {
	fmt.Println("step 1")
	tenant, err := s.tenantStore.GetTenantById(tenantId)
	if err != nil {
		return "", err
	}
	if tenant.State != tenants.Active {
		return "", ErrTenantIsNotValid
	}
	fmt.Println("step 2")
	existing, err := s.apiKeyStore.GetHashedAPIKeyByNameAndTenantID(name, tenantId)
	if err != nil && err != ErrKeyNotFound {
		return "", err
	}
	if existing.ID > 0 {
		return "", ErrKeyNameTenantIDCombinationNotUnique
	}

	fmt.Println("step 3")
	newApiKey, err := newApiKey(name, expirationDate)
	if err != nil {
		return "", err
	}
	hashed, err := newApiKey.hash()
	if err != nil {
		return "", err
	}
	fmt.Println("step 4")
	err = s.apiKeyStore.AddApiKey(tenant.ID, hashed)
	if err != nil {
		return "", err
	}
	fmt.Println("step 5")
	apiKey := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(
		[]byte(fmt.Sprintf("%d:%s", newApiKey.ID, newApiKey.Secret)))
	fmt.Println("step 6")
	return apiKey, nil
}

// Authenticates a given API key. Input must be 'apiKeyId:apiKey' encoded to a base64 string
// API key is valid if it is the correct api key id and api key combination and if the attached tenant is active
func (s *Service) AuthenticateApiKey(base64IdAndKeyCombination string) (ApiKeyAuthenticationDTO, error) {
	apiKeyId, apiKey, err := apiKeyAndIdFromBase64(base64IdAndKeyCombination)
	if err != nil {
		return ApiKeyAuthenticationDTO{}, ErrInvalidEncoding
	}
	hashed, err := s.apiKeyStore.GetHashedApiKeyById(apiKeyId, []tenants.State{tenants.Active})
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
		if hashed.ExpirationDate != nil {
			exp := hashed.ExpirationDate.Unix()
			return ApiKeyAuthenticationDTO{
				TenantID:   fmt.Sprintf("%d", hashed.TenantID),
				Expiration: &exp,
			}, nil
		} else {
			return ApiKeyAuthenticationDTO{
				TenantID: fmt.Sprintf("%d", hashed.TenantID),
			}, nil
		}
	}
	return ApiKeyAuthenticationDTO{}, ErrKeyNotFound
}

type Filter struct {
	TenantID []int64 `schema:"tenant_id"`
}

type ApiKeyAuthenticationDTO struct {
	TenantID   string `json:"sub"` // Sub is how Ory Oathkeeper identifies the important information in the response
	Expiration *int64 `json:"expiration_date"`
}

type ApiKeyDTO struct {
	ID             int64      `json:"id"`
	Name           string     `json:"name"`
	TenantID       int64      `json:"tenant_id"`
	TenantName     string     `json:"tenant_name"`
	ExpirationDate *time.Time `json:"expiration_date"`
	Created        time.Time  `json:"created"`
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

type Service struct {
	tenantStore tenantStore
	apiKeyStore apiKeyStore
}

type apiKeyStore interface {
	AddApiKey(tenantID int64, hashedApiKey HashedApiKey) error
	DeleteApiKey(id int64) error
	GetHashedApiKeyById(id int64, stateFilter []tenants.State) (HashedApiKey, error)
	GetHashedAPIKeyByNameAndTenantID(name string, tenantID int64) (HashedApiKey, error)
	List(Filter, pagination.Request) (*pagination.Page[ApiKeyDTO], error)
}

type tenantStore interface {
	GetTenantById(id int64) (tenants.Tenant, error)
}
