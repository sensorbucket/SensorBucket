package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrMalformedJWT = errors.New("JWT is malformed")
	ErrJWTExpired   = errors.New("JWT is expired")
	ErrNoAPIKey     = errors.New("no APIKey available")
)

type APIKeyTrader func(jwt string) (string, error)

type ClientKeySets struct {
	apiKeys  map[string]string
	jwts     *KV[string]
	tradeKey APIKeyTrader
}

func CreateClientKeySets(ctx context.Context, apiKeyTrader APIKeyTrader) *ClientKeySets {
	ks := &ClientKeySets{
		apiKeys:  map[string]string{},
		jwts:     NewKV[string](ctx),
		tradeKey: apiKeyTrader,
	}
	go ks.jwts.StartCleaner(5 * time.Second)
	return ks
}

func (keysets *ClientKeySets) Authenticate(clientID, apiKey string) error {
	jwt, err := keysets.tradeKey(apiKey)
	if err != nil {
		return err
	}

	// APIKey can be traded for JWT so its valid
	keysets.apiKeys[clientID] = apiKey
	if err := keysets.updateJWT(clientID, jwt); err != nil {
		return fmt.Errorf("updating JWT in memory: %w", err)
	}

	return nil
}

func (keysets *ClientKeySets) Destroy(clientID string) {
	delete(keysets.apiKeys, clientID)
	keysets.jwts.Delete(clientID)
}

func (keysets *ClientKeySets) GetClientJWT(clientID string) (string, error) {
	jwt, ok := keysets.jwts.Get(clientID)
	if ok {
		return jwt, nil
	}

	return keysets.refreshJWT(clientID)
}

func (keysets *ClientKeySets) refreshJWT(clientID string) (string, error) {
	apiKey, ok := keysets.apiKeys[clientID]
	if !ok {
		return "", ErrNoAPIKey
	}

	jwt, err := keysets.tradeKey(apiKey)
	if err != nil {
		return "", fmt.Errorf("authenticating APIKey: %w", err)
	}
	if err := keysets.updateJWT(clientID, jwt); err != nil {
		return "", fmt.Errorf("updating JWT in memory: %w", err)
	}

	return jwt, nil
}

type jwtExpiry struct {
	Expiry int64 `json:"exp"`
}

func (keysets *ClientKeySets) updateJWT(clientID, jwt string) error {
	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return ErrMalformedJWT
	}

	data, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return err
	}

	var body jwtExpiry
	if err := json.Unmarshal(data, &body); err != nil {
		return err
	}

	expiry := time.UnixMilli(body.Expiry)
	if time.Now().After(expiry) {
		return ErrJWTExpired
	}
	keysets.jwts.Set(clientID, jwt, expiry)

	return nil
}
