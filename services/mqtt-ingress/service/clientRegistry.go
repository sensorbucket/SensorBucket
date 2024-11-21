package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrMalformedJWT   = errors.New("JWT is malformed")
	ErrJWTExpired     = errors.New("JWT is expired")
	ErrClientNotFound = errors.New("client info is missing")
)

type APIKeyTrader func(jwt string) (string, error)

type ClientInfo struct {
	// Static; won't be refreshed
	PipelineID uuid.UUID
	APIKey     string
	// Dynamic; will be refreshed
	TenantID    int64
	AccessToken string
	Expiry      time.Time
}

// Client manages ClientInfo and performs a concurrent safe refresh of the access token after expiry
type Client struct {
	info         ClientInfo
	tradeKey     APIKeyTrader
	jwtTTLMargin time.Duration
	lock         sync.RWMutex
}

func NewClient(pipelineID uuid.UUID, apiKey string, tradeKey APIKeyTrader, jwtTTLMargin time.Duration) (*Client, error) {
	client := &Client{
		info: ClientInfo{
			PipelineID: pipelineID,
			APIKey:     apiKey,
		},
		tradeKey:     tradeKey,
		jwtTTLMargin: jwtTTLMargin,
		lock:         sync.RWMutex{},
	}
	if err := client.refresh(); err != nil {
		return nil, err
	}
	return client, nil
}

func (client *Client) refresh() error {
	jwt, err := client.tradeKey(client.info.APIKey)
	if err != nil {
		return fmt.Errorf("authenticating APIKey: %w", err)
	}
	claims, err := extractClaims(jwt)
	if err != nil {
		return fmt.Errorf("extracting JWT claims: %w", err)
	}
	expiry := time.Unix(claims.Expiry, 0).Add(-client.jwtTTLMargin)
	if time.Now().After(expiry) {
		return fmt.Errorf("%w at %v, TTLMargin too big", ErrJWTExpired, expiry)
	}
	log.Printf("MQTT Client (%s) refreshed, expires at: %s\n", "unknown", expiry.String())

	client.info.AccessToken = jwt
	client.info.TenantID = claims.TenantID
	client.info.Expiry = expiry
	return nil
}

func (client *Client) Info() (ClientInfo, error) {
	client.lock.RLock()
	info := client.info
	if time.Now().Before(info.Expiry) {
		client.lock.RUnlock()
		return info, nil
	}
	client.lock.RUnlock()

	// Info was expired when we read it, take full lock and refresh info
	// Make sure that the info wasn't refreshed in between our locks
	client.lock.Lock()
	// Info was refreshed between locks, return new info
	if time.Now().Before(client.info.Expiry) {
		info = client.info
		client.lock.Unlock()
		return info, nil
	}

	if err := client.refresh(); err != nil {
		return ClientInfo{}, err
	}
	info = client.info
	client.lock.Unlock()

	return info, nil
}

// ClientRegistry is a glorified slice with CLient structs mapped using their clientID
type ClientRegistry struct {
	clients      map[string]*Client
	jwtTTLMargin time.Duration
	tradeKey     APIKeyTrader
}

func CreateClientRegistry(ctx context.Context, apiKeyTrader APIKeyTrader) *ClientRegistry {
	ks := &ClientRegistry{
		clients:      map[string]*Client{},
		tradeKey:     apiKeyTrader,
		jwtTTLMargin: time.Minute * 1,
	}
	return ks
}

func (registry *ClientRegistry) Authenticate(clientID, username, apiKey string) error {
	pipelineID, err := uuid.Parse(username)
	if err != nil {
		return fmt.Errorf("invalid pipelineID: %w", err)
	}

	client, err := NewClient(pipelineID, apiKey, registry.tradeKey, registry.jwtTTLMargin)
	if err != nil {
		return err
	}

	registry.Destroy(clientID)
	registry.clients[clientID] = client

	return nil
}

func (registry *ClientRegistry) Destroy(clientID string) {
	delete(registry.clients, clientID)
}

func (registry *ClientRegistry) GetClient(clientID string) (ClientInfo, error) {
	client, ok := registry.clients[clientID]
	if !ok {
		return ClientInfo{}, ErrClientNotFound
	}

	return client.Info()
}

type jwtClaims struct {
	Expiry   int64 `json:"exp"`
	TenantID int64 `json:"tid"`
}

func extractClaims(jwt string) (jwtClaims, error) {
	var claims jwtClaims

	parts := strings.Split(jwt, ".")
	if len(parts) != 3 {
		return claims, ErrMalformedJWT
	}

	data, err := base64.StdEncoding.WithPadding(base64.NoPadding).DecodeString(parts[1])
	if err != nil {
		return claims, err
	}

	if err := json.Unmarshal(data, &claims); err != nil {
		return claims, err
	}

	expiry := time.Unix(claims.Expiry, 0)
	if time.Now().After(expiry) {
		return claims, fmt.Errorf("%w at %v", ErrJWTExpired, expiry)
	}

	return claims, nil
}
