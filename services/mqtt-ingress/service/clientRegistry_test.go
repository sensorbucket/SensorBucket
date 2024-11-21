package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sensorbucket.nl/sensorbucket/pkg/authtest"
	"sensorbucket.nl/sensorbucket/services/mqtt-ingress/service"
)

func TestShouldDestroyPreviousSession(t *testing.T) {
	tradeCalls := 0
	registry := service.CreateClientRegistry(context.Background(), func(apikey string) (string, error) {
		tradeCalls++
		return authtest.CreateToken(), nil
	}, 0)
	clientID := "asession123"
	require.NoError(t, registry.Authenticate(clientID, "db2053d9-90bd-4b83-b844-07dc51ba7591", "apikey123"))
	info, err := registry.GetClient(clientID)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	require.NoError(t, registry.Authenticate(clientID, "db2053d9-90bd-4b83-b844-07dc51ba7591", "apikey123"))
	info2, err := registry.GetClient(clientID)
	assert.NoError(t, err)
	assert.NotEqual(t, info2.AccessToken, info.AccessToken)
	assert.Equal(t, 2, tradeCalls)
}

func TestShouldRefreshOnJWTExpire(t *testing.T) {
	tradeCalls := 0
	registry := service.CreateClientRegistry(context.Background(), func(apikey string) (string, error) {
		tradeCalls++
		return authtest.CreateTokenWithExpiry(time.Now().Add(1 * time.Second)), nil
	}, 0)
	clientID := "asession123"
	require.NoError(t, registry.Authenticate(clientID, "db2053d9-90bd-4b83-b844-07dc51ba7591", "apikey123"))
	_, err := registry.GetClient(clientID)
	assert.NoError(t, err)
	time.Sleep(1 * time.Second)
	_, err = registry.GetClient(clientID)
	assert.NoError(t, err)
	assert.Equal(t, 2, tradeCalls)
}

// TestShouldPropogateTradeError also applies to authentication errors, since the keyTrader would error with 401
func TestShouldPropogateTradeError(t *testing.T) {
	registry := service.CreateClientRegistry(context.Background(), func(apikey string) (string, error) {
		return "", errors.New("some funky (fake because this is a test) http error")
	}, 0)
	clientID := "asession123"
	require.Error(t, registry.Authenticate(clientID, "db2053d9-90bd-4b83-b844-07dc51ba7591", "apikey123"))
}

func TestShouldPropogateTradeErrorAfterAuthenticate(t *testing.T) {
	tradeCalls := 0
	registry := service.CreateClientRegistry(context.Background(), func(apikey string) (string, error) {
		tradeCalls++
		if tradeCalls == 1 {
			return authtest.CreateTokenWithExpiry(time.Now().Add(1 * time.Second)), nil
		}
		return "", errors.New("some funky (fake because this is a test) http error")
	}, 0)
	clientID := "asession123"
	require.NoError(t, registry.Authenticate(clientID, "db2053d9-90bd-4b83-b844-07dc51ba7591", "apikey123"))
	time.Sleep(1 * time.Second)
	_, err := registry.GetClient(clientID)
	assert.Error(t, err)
	assert.Equal(t, 2, tradeCalls)
}
