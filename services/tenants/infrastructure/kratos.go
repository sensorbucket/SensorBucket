package tenantsinfra

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	ory "github.com/ory/client-go"

	"sensorbucket.nl/sensorbucket/services/tenants/tenants"
)

var _ tenants.UserValidator = (*KratosUserValidator)(nil)

type KratosUserValidator struct {
	client *ory.APIClient
}

func NewKratosUserValidator(adminURL string) *KratosUserValidator {
	cfg := ory.NewConfiguration()
	cfg.Servers = []ory.ServerConfiguration{
		{
			URL: adminURL,
		},
	}
	client := ory.NewAPIClient(cfg)
	return &KratosUserValidator{
		client: client,
	}
}

func (kratos *KratosUserValidator) UserByIDExists(ctx context.Context, tenantID int64, userID string) error {
	identity, httpRes, err := kratos.client.IdentityAPI.GetIdentity(ctx, userID).Execute()
	if httpRes.StatusCode == http.StatusNotFound {
		return errors.New("user not found")
	}
	if err != nil {
		return fmt.Errorf("in UserByIDExists, couldn't get identity from kratos admin api: %w", err)
	}
	if identity.GetState() != ory.IDENTITYSTATE_ACTIVE {
		return fmt.Errorf("in UserByIDExists, identity is not active but '%s'", identity.GetState())
	}
	return nil
}
