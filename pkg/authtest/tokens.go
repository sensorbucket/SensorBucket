package authtest

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"

	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var (
	DefaultTenantID int64  = 10
	DefaultSub      string = "ONLYFORTESTING"
)

func GodContext() context.Context {
	return auth.CreateAuthenticatedContextForTESTING(context.Background(), DefaultSub, DefaultTenantID, auth.AllPermissions())
}

func CreateToken() string {
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"tid":   DefaultTenantID,
		"perms": auth.AllPermissions(),
		"sub":   DefaultSub,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	token.Header["kid"] = "test-key"
	tokenString, err := token.SignedString(JsonPrivateKey())
	if err != nil {
		panic(err)
	}
	return tokenString
}

func AuthenticateRequest(req *http.Request) {
	req.Header.Add("Authorization", "Bearer "+CreateToken())
}

type StaticJWKSProvider jose.JSONWebKeySet

func (jwks StaticJWKSProvider) Get() (jose.JSONWebKeySet, error) {
	return jose.JSONWebKeySet(jwks), nil
}

func JWKS() StaticJWKSProvider {
	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(io.NopCloser(strings.NewReader(jsonWebKeySet))).Decode(&jwks); err != nil {
		return StaticJWKSProvider{}
	}
	return StaticJWKSProvider(jwks)
}

func JsonPrivateKey() any {
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
