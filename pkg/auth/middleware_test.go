package auth

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
)

// test jwks is unreachable
func TestAuthenticateWellKnownUnreachable(t *testing.T) {
	// Arrange
	client := jwksClientMock{
		GetFunc: func() (jose.JSONWebKeySet, error) {
			return jose.JSONWebKeySet{}, fmt.Errorf("connection refused")
		},
	}
	auth := Authenticate(&client)
	s := http.ServeMux{}
	s.Handle("/", auth(nil))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	token := createToken(jwt.MapClaims{
		"tid": 11,
		"perms": []string{
			"READ_DEVICES",
			"READ_API_KEYS",
		},
		"uid": 431,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Act
	s.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, 401, rr.Result().StatusCode)
}

func TestProtectAndAuthenticatePassClaimsToNext(t *testing.T) {
	// Arrange
	protect := Protect()
	client := jwksClientMock{
		GetFunc: func() (jose.JSONWebKeySet, error) {
			return jwks(), nil
		},
	}
	auth := Authenticate(&client)
	next := &HandlerMock{
		ServeHTTPFunc: func(_ http.ResponseWriter, r *http.Request) {
			tID, err := GetTenant(r.Context())
			assert.NoError(t, err)
			assert.Equal(t, int64(11), tID)

			uID, err := GetUser(r.Context())
			assert.NoError(t, err)
			assert.Equal(t, int64(431), uID)

			permissions, err := GetPermissions(r.Context())
			assert.NoError(t, err)
			assert.Equal(t, Permissions{READ_DEVICES, READ_API_KEYS}, permissions)
		},
	}
	s := http.ServeMux{}
	s.Handle("/", auth(protect(next)))

	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	token := createToken(jwt.MapClaims{
		"tid": 11,
		"perms": []string{
			"READ_DEVICES",
			"READ_API_KEYS",
		},
		"uid": 431,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
	})
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// Act
	s.ServeHTTP(rr, req)

	// Assert
	assert.Equal(t, 200, rr.Result().StatusCode)
	assert.Len(t, next.ServeHTTPCalls(), 1)
}

func TestProtect(t *testing.T) {
	type testCase struct {
		values             map[ctxKey]any
		expectedStatusCode int
		expectedNextCalls  int
	}
	scenarios := map[string]testCase{
		"all required values present": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(11),
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 200,
			expectedNextCalls:  1,
		},
		"tid is missing": {
			values: map[ctxKey]any{
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"perms is missing": {
			values: map[ctxKey]any{
				ctxTenantID: int64(12),
				ctxUserID:   int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"uid is missing": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(13),
				ctxPermissions: Permissions{READ_API_KEYS},
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"all required values are missing": {
			values:             map[ctxKey]any{},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"tid is wrong type": {
			values: map[ctxKey]any{
				ctxTenantID:    "123", // should be int64!
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"perms is wrong type": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(13),
				ctxPermissions: 54325,
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"uid is wrong type": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(13),
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      "asdasdsad",
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"tid is nil": {
			values: map[ctxKey]any{
				ctxTenantID:    nil,
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"perms is nil": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(13),
				ctxPermissions: nil,
				ctxUserID:      int64(124),
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"uid is nil": {
			values: map[ctxKey]any{
				ctxTenantID:    int64(13),
				ctxPermissions: Permissions{READ_API_KEYS},
				ctxUserID:      nil,
			},
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			ctx := createTestContext(context.Background(), cfg.values)

			next := &HandlerMock{
				ServeHTTPFunc: func(responseWriter http.ResponseWriter, request *http.Request) {},
			}

			handler := Protect()
			s := http.ServeMux{}
			s.Handle("/", handler(next))

			// Act
			s.ServeHTTP(rr, req.WithContext(ctx))

			// Assert
			assert.Equal(t, cfg.expectedStatusCode, rr.Result().StatusCode)
			assert.Len(t, next.ServeHTTPCalls(), cfg.expectedNextCalls)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	in24Hours := time.Now().Add(time.Hour * 24).Unix()
	type testCase struct {
		authHeader          string
		expectedStatusCode  int
		expectedNextCalls   int
		expectedTenantID    *int64
		expectedUserID      *int64
		expectedPermissions Permissions
	}
	scenarios := map[string]testCase{
		"auth header is invalid": {
			authHeader:         "blabla",
			expectedStatusCode: 400,
			expectedNextCalls:  0,
		},
		"bearer token is invalid": {
			authHeader:         "Bearer blabla",
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"anonymous request is done": {
			authHeader:         "",
			expectedStatusCode: 200,
			expectedNextCalls:  1,
		},
		"bearer token is valid and contains all claims": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"tid": 11,
					"perms": []string{
						"READ_DEVICES",
						"READ_API_KEYS",
					},
					"uid": 431,
					"exp": in24Hours,
				},
			)),
			expectedStatusCode:  200,
			expectedNextCalls:   1,
			expectedUserID:      lo.ToPtr[int64](431),
			expectedTenantID:    lo.ToPtr[int64](11),
			expectedPermissions: Permissions{READ_DEVICES, READ_API_KEYS},
		},
		"bearer token contains invalid permission": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"tid": 11,
					"perms": []string{
						"READ_DEVICES",
						"READ_API_KEYS",
						"DOES_NOT_EXIST",
					},
					"uid": 431,
					"exp": in24Hours,
				},
			)),
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"bearer token is valid and but claims are missing": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{},
			)),
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"bearer token is valid but tid is missing": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"perms": []string{
						"READ_DEVICES",
						"READ_API_KEYS",
					},
					"uid": 431,
					"exp": in24Hours,
				},
			)),
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"bearer token is valid but perms is missing": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"tid": 11,
					"uid": 431,
					"exp": in24Hours,
				},
			)),
			expectedStatusCode: 200,
			expectedNextCalls:  1,
			expectedTenantID:   lo.ToPtr[int64](11),
			expectedUserID:     lo.ToPtr[int64](431),
		},
		"bearer token is valid but uid is missing": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"tid": 11,
					"perms": []string{
						"READ_DEVICES",
						"READ_API_KEYS",
					},
					"exp": in24Hours,
				},
			)),
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
		"bearer token is valid all claims are present but the token is expired": {
			authHeader: fmt.Sprintf("Bearer %s", createToken(
				jwt.MapClaims{
					"tid": 11,
					"perms": []string{
						"READ_DEVICES",
						"READ_API_KEYS",
					},
					"uid": 431,
					"exp": time.Now().Add(-time.Hour * 24).Unix(),
				},
			)),
			expectedStatusCode: 401,
			expectedNextCalls:  0,
		},
	}

	for scene, cfg := range scenarios {
		t.Run(scene, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Authorization", cfg.authHeader)
			rr := httptest.NewRecorder()

			next := HandlerMock{
				ServeHTTPFunc: func(_ http.ResponseWriter, r *http.Request) {
					tID, err := GetTenant(r.Context())
					if cfg.expectedTenantID == nil {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, *cfg.expectedTenantID, tID)
					}

					uID, err := GetUser(r.Context())
					if cfg.expectedTenantID == nil {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, *cfg.expectedUserID, uID)
					}

					permissions, err := GetPermissions(r.Context())
					if cfg.expectedTenantID == nil {
						assert.Error(t, err)
					} else {
						assert.NoError(t, err)
						assert.Equal(t, cfg.expectedPermissions, permissions)
					}
				},
			}

			client := jwksClientMock{
				GetFunc: func() (jose.JSONWebKeySet, error) {
					return jwks(), nil
				},
			}

			handler := Authenticate(&client)
			s := http.ServeMux{}
			s.Handle("/", handler(&next))

			// Act
			s.ServeHTTP(rr, req)

			// Assert
			assert.Equal(t, cfg.expectedStatusCode, rr.Result().StatusCode)
			assert.Len(t, next.ServeHTTPCalls(), cfg.expectedNextCalls)
		})
	}
}

func jsonPrivateKey() any {
	block, _ := pem.Decode([]byte(key))
	if block == nil {
		panic("failed to parse PEM block containing the private key")
	}
	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	return privateKey
}

func createTestContext(ctx context.Context, values map[ctxKey]any) context.Context {
	for key, val := range values {
		ctx = context.WithValue(ctx, key, val)
	}
	return ctx
}

func createToken(claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(jsonPrivateKey())
	if err != nil {
		panic(err)
	}
	return tokenString
}

func jwks() jose.JSONWebKeySet {
	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(io.NopCloser(strings.NewReader(jsonWebKeySet))).Decode(&jwks); err != nil {
		return jose.JSONWebKeySet{}
	}
	return jwks
}

// Keys below are for testing purposes only!!
const key = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDNoUGP4FABzt8m
XO/uoSrQ/thVTHDG2Lb3pQLmC6BkgCzygBtO8eTeORNkQHirNKC47yk8mllF2RdJ
doHiDFyfRSa+V8AJv4KfF7Yb65J8a78yAcmnj6yTSQqM+7E2U7WMRTbYw9HyE4Zp
Xp42pPCGFArlsT6CkPSAL+eLIvVjbSsv71DIy6UsDsRuAiK+27JC9FEjqJst9fLk
QXqawC8gNeE0lXN91Wj62sP1a9i7D+MFD3p92UI2F3FNilOKrCPrsQI9Y5Il9qHK
u+HM3AJ//7Ym3/RN69jsBBZAclRaCiBlJhhczMkzxUffiLkxe1hiNhTZBOm7n0a2
MmFyqDLNAgMBAAECggEAHXODquIzQ1cIUfvMp45wzfc6L9lfa7N9XTHGnQE8Sziq
d18OyjtODt/43Yp4XfkPLf2fF915fM4PjkeJacFggLVMS8XQrPS/dh7Ux+HxHJ3o
B/cGlVe4HW5AMxoXcxMBNSJyrRA64SOXxD63hVcRVfrH5scAj33IbxWtYZmzsLYf
1/TWaY5DEd3i67W65tDNzSVoCYu8Wsg5z6lmN5SJmxR1zjMyypCoGNdcm9Pa/vvq
Hb2xHKOX3Io4vSY2VTurWk9/iIfEVLuqiuq1s5dJ2vd3OCHslw2JshOGM+kGU9Lt
z6+lcBJcr7jPAPL8y4EMgs1oqsNBfUIXkr59UrYyZwKBgQDnwv3UHFvuTvKgAdqj
UW8fxuWoJ46KBBNAVCkuO8RNHpoFG5dsfHom8hLMPi0d/+9udN6k4Aac4AhpV4at
RFKjdHjBVsm06TKSf8fPGUselicWCBuqUFHH3Pi20Aw+i9R5aAxxYR12gOiZJW0D
eLxnHVKgtVwDFfg3JAR0dOik/wKBgQDjIqHrxh98mshdYUSmWFYYjCSONCsIoGmo
DVdl8LKfNgMzegRCcKjteUjESXipm5Z2uitiQWNpe0HR4bltCyAyzkenTBqOf+Yh
TfJTB94ko7RR22Xj71WeI9WRnCOINQXIvHNBSf3gYXccZjBV20cr9xEKq0qjS6YV
ZffCYoysMwKBgDjUQXVvdsNarHe7vKbrYvpBxTKUcIk7MpVFjct+cEYQyOeTum+p
njJKjX1ziZCfn1BQa/+1xylUbfuWsLlv1WurNakC5Pbtb68okhAgPaFEZFUsq8v5
YfRGJN5+6WG02+bhMpvimlzigyZ6XN7LDjeiow4xKly/WFv9AvKjcCB1AoGAGThm
PFTSeDaDmwLK6aGTZcRh5rxaLuoI8VUR6ErSuqT3tAaPZIU37K5z6v+xezvAeExx
tsZF8Jd0Fob23OnIWHvZLvVfWYVQG1CZYKjV/MGEqzYuWSHhIt8dvr5Un7Irgz+R
mKVLoFeSL0AVi+L+Qx568PFWJ02mEmgxG49vyUsCgYBm6R13DaGv5mylpYc/CWbx
rF3IpRWYewlcO2xrgiCEvp+9Eh0epSuK/kKaEwwv90pMHReIrpcMujBOpUJT7/NZ
fJA0UGp5r4Z2az1b4i4sF70Uark9TatJ3XH7AcP3tFfo2TQeiST4qgKyx35iT/0r
mxiuHhps1ig5jCN3YGj2zQ==
-----END PRIVATE KEY-----
`

//const publicKey = `-----BEGIN PUBLIC KEY-----
//MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAzaFBj+BQAc7fJlzv7qEq
//0P7YVUxwxti296UC5gugZIAs8oAbTvHk3jkTZEB4qzSguO8pPJpZRdkXSXaB4gxc
//n0UmvlfACb+Cnxe2G+uSfGu/MgHJp4+sk0kKjPuxNlO1jEU22MPR8hOGaV6eNqTw
//hhQK5bE+gpD0gC/niyL1Y20rL+9QyMulLA7EbgIivtuyQvRRI6ibLfXy5EF6msAv
//IDXhNJVzfdVo+trD9WvYuw/jBQ96fdlCNhdxTYpTiqwj67ECPWOSJfahyrvhzNwC
//f/+2Jt/0TevY7AQWQHJUWgogZSYYXMzJM8VH34i5MXtYYjYU2QTpu59GtjJhcqgy
//zQIDAQAB
//-----END PUBLIC KEY-----
//`

const jsonWebKeySet = `{
	"keys":[
	   {
		  "alg":"RS256",
		  "e":"AQAB",
		  "kid":"test-key",
		  "kty":"RSA",
		  "n":"zaFBj-BQAc7fJlzv7qEq0P7YVUxwxti296UC5gugZIAs8oAbTvHk3jkTZEB4qzSguO8pPJpZRdkXSXaB4gxcn0UmvlfACb-Cnxe2G-uSfGu_MgHJp4-sk0kKjPuxNlO1jEU22MPR8hOGaV6eNqTwhhQK5bE-gpD0gC_niyL1Y20rL-9QyMulLA7EbgIivtuyQvRRI6ibLfXy5EF6msAvIDXhNJVzfdVo-trD9WvYuw_jBQ96fdlCNhdxTYpTiqwj67ECPWOSJfahyrvhzNwCf_-2Jt_0TevY7AQWQHJUWgogZSYYXMzJM8VH34i5MXtYYjYU2QTpu59GtjJhcqgyzQ"
	   }
	]
 }`
