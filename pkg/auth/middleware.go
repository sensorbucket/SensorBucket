package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"

	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/api"
)

type claims struct {
	jwt.StandardClaims
	TenantID    int64       `json:"tid"`
	Permissions Permissions `json:"perms"`
}

func (c *claims) Valid() error {
	if err := c.Permissions.Validate(); err != nil {
		return err
	}
	if c.ExpiresAt > time.Now().Unix() {
		return nil
	}
	return fmt.Errorf("claims not valid")
}

type JWKSClient interface {
	Get() (jose.JSONWebKeySet, error)
}

type jwksHttpClient struct {
	url        string
	httpClient http.Client
}

func (c *jwksHttpClient) Get() (jose.JSONWebKeySet, error) {
	res, err := c.httpClient.Get(c.url)
	if err != nil {
		return jose.JSONWebKeySet{}, fmt.Errorf("failed to fetch jwks: %w", err)
	}
	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		return jose.JSONWebKeySet{}, fmt.Errorf("failed to decode jwks: %w", err)
	}
	return jwks, nil
}

func NewJWKSHttpClient(url string) *jwksHttpClient {
	return &jwksHttpClient{
		url:        url,
		httpClient: http.Client{},
	}
}

func Protect() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := GetTenant(r.Context()); err != nil {
				log.Printf("[Auth] getting tenant: %v\n", err)
				web.HTTPError(w, ErrUnauthorized)
				return
			}
			if _, err := GetUser(r.Context()); err != nil && !errors.Is(err, ErrContextMissing) {
				log.Printf("[Auth] getting user: %v\n", err)
				web.HTTPError(w, ErrUnauthorized)
				return
			}
			if _, err := GetPermissions(r.Context()); err != nil {
				log.Printf("[Auth] getting permissions: %v\n", err)
				web.HTTPError(w, ErrUnauthorized)
				return
			}
			// All required authentication values are present, allow the request
			next.ServeHTTP(w, r)
		})
	}
}

func ForwardRequestAuthentication() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token, ok := StripBearer(r.Header.Get("Authorization"))
			if ok {
				r = r.WithContext(context.WithValue(
					r.Context(), api.ContextAccessToken, token,
				))
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Authentication middleware for checking the validity of any present JWT
// Checks if the JWT is signed using the given secret
// Serves the next HTTP handler if there is no JWT or if the JWT is OK
// Anonymous requests are allowed by this handler
func Authenticate(keyClient JWKSClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authStr := r.Header.Get("Authorization")
			if authStr == "" {
				// Allow anonymous requests
				next.ServeHTTP(w, r)
				return
			}

			// Cheating, removes Bearer and bearer case independently
			tokenStr, ok := StripBearer(authStr)
			if !ok {
				log.Printf("[Error] authentication failed err because the Authorization header is malformed\n")
				web.HTTPError(w, ErrAuthHeaderInvalidFormat)
				return
			}

			// Retrieve the JWT and ensure it was signed by us
			c := claims{}
			token, err := jwt.ParseWithClaims(tokenStr, &c, validateJWTFunc(keyClient))
			if err != nil {
				log.Printf("[Error] authentication failed err: %s\n", err)
				web.HTTPError(w, ErrUnauthorized)
				return
			}
			if !token.Valid {
				log.Printf("[Error] authentication failed err: %s\n", err)
				web.HTTPError(w, ErrUnauthorized)
				return
			}
			// JWT itself is validated, pass it to the actual endpoint for further authorization
			// First fill the context with user information
			ctx := setTenantID(r.Context(), c.TenantID)
			ctx = setUserID(ctx, c.Subject)
			ctx = setPermissions(ctx, c.Permissions)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateJWTFunc(jwksClient JWKSClient) func(token *jwt.Token) (any, error) {
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
