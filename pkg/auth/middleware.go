package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

type ctxKey int

type claims struct {
	TenantID    int64        `json:"tid"`
	Permissions []permission `json:"perms"`
	UserID      int64        `json:"uid"`
	Expiration  int64        `json:"exp"`
}

func (c *claims) Valid() error {
	for _, permission := range c.Permissions {
		if permission.Valid() != nil {
			return fmt.Errorf("invalid permissions")
		}
	}
	if c.TenantID > 0 && c.UserID > 0 && c.Expiration > time.Now().Unix() {
		return nil
	}
	return fmt.Errorf("claims not valid")
}

type jwksClient interface {
	Get() (jose.JSONWebKeySet, error)
}

type jwksHttpClient struct {
	issuer     string
	httpClient http.Client
}

func (c *jwksHttpClient) Get() (jose.JSONWebKeySet, error) {
	res, err := c.httpClient.Get(fmt.Sprintf("%s/.well-known/jwks.json", c.issuer))

	if err != nil {
		return jose.JSONWebKeySet{}, fmt.Errorf("failed to fetch jwks: %w", err)
	}
	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		return jose.JSONWebKeySet{}, fmt.Errorf("failed to decode jwks: %w", err)
	}
	return jwks, nil
}

type contextBuilder struct {
	c context.Context
}

func (cb *contextBuilder) With(key ctxKey, value any) *contextBuilder {
	cb.c = context.WithValue(cb.c, key, value)
	return cb
}

func (cb *contextBuilder) Finish() context.Context {
	return cb.c
}

const (
	ctxUserID ctxKey = iota
	ctxCurrentTenantID
	ctxPermissions
)

func NewJWKSHttpClient(issuer string) *jwksHttpClient {
	return &jwksHttpClient{
		issuer:     issuer,
		httpClient: http.Client{},
	}
}

func Protect() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, tenantIDPresent := fromRequestContext[[]int64](r.Context(), ctxCurrentTenantID)
			_, permissionsPresent := fromRequestContext[[]permission](r.Context(), ctxPermissions)
			_, userIDPresent := fromRequestContext[int64](r.Context(), ctxUserID)
			if tenantIDPresent && permissionsPresent && userIDPresent {
				// All required authentication values are present, allow the request
				next.ServeHTTP(w, r)
				return
			}
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

// Authentication middleware for checking the validity of any present JWT
// Checks if the JWT is signed using the given secret
// Serves the next HTTP handler if there is no JWT or if the JWT is OK
// Anonymous requests are allowed by this handler
func Authenticate(keyClient jwksClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				// Allow anonymous requests
				next.ServeHTTP(w, r)
				return
			}
			tokenStr, ok := strings.CutPrefix(auth, "Bearer ")
			if !ok {
				web.HTTPError(w, ErrAuthHeaderInvalidFormat)
				return
			}

			// Retrieve the JWT and ensure it was signed by us
			c := claims{}
			token, err := jwt.ParseWithClaims(tokenStr, &c, validateJWTFunc(keyClient))
			if err == nil && token.Valid {
				// JWT itself is validated, pass it to the actual endpoint for further authorization
				// First fill the context with user information
				cb := contextBuilder{c: r.Context()}
				next.ServeHTTP(w, r.WithContext(
					cb.
						With(ctxCurrentTenantID, []int64{c.TenantID}).
						With(ctxUserID, c.UserID).
						With(ctxPermissions, c.Permissions).
						Finish()))
				return
			}
			log.Printf("[Error] authentication failed err: %s", err)
			web.HTTPError(w, ErrUnauthorized)
		})
	}
}

func validateJWTFunc(jwksClient jwksClient) func(token *jwt.Token) (any, error) {
	return func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Retrieve JWKS
		jwks, err := jwksClient.Get()
		if err != nil {
			return nil, fmt.Errorf("failed to retrieve jwks: %w", err)
		}

		// Look for the key as indicated by the token key id
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, fmt.Errorf("no kid in token")
		}
		keys := jwks.Key(kid)
		if len(keys) == 0 {
			return nil, fmt.Errorf("no keys found for token")
		}
		key := keys[0]
		if key.Algorithm != token.Method.Alg() {
			return nil, fmt.Errorf("key alg differs from token alg: %s vs %s", key.Algorithm, token.Method.Alg())
		}
		return key.Public().Key, nil
	}
}
