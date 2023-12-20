package tenants

import "fmt"

type Service struct{}

var (
	ErrTenantIsNotValid = fmt.Errorf("Tenant is not valid")
	ErrKeyNotFound      = fmt.Errorf("couldnt find key")
)

func (s *Service) GetTenantById(id int64) (Tenant, error) {
	return Tenant{}, nil
}

func (s *Service) GetHashedApiKeyById(id int64) (ApiKey, error) {
	return ApiKey{}, nil
}

func (s *Service) GenerateNewApiKey(owner Tenant) (string, error) {
	return "", nil
}
