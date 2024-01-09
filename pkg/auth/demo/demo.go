package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-jose/go-jose/v3"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var secretKey = []byte("d0n1 wu3ry, ve3y s4f3")

func main() {
	r := chi.NewRouter()

	// Use middleware to validate JWT
	r.Use(auth.Authenticate("http://localhost:4467"))
	r.Use(auth.Protect())

	t, err := createToken()
	if err != nil {
		panic(err)
	}
	fmt.Println("TOKEN", t)

	// Define your routes
	r.Get("/stuff", someProtectedEndpoint())

	go http.ListenAndServe(":8089", r)

	fmt.Println("Sleeping for 1s...")
	time.Sleep(time.Second * 1)

	c := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8089/stuff", nil)
	req.Header.Set("Authorization", "Bearer "+t)
	if err != nil {
		panic(err)
	}
	resp, err := c.Do(req)
	if err != nil {
		panic(err)
	}
	fmt.Println("resp", resp, resp.StatusCode)

	for {
	}
}

type Filter struct {
	Tenants []int64
}

func someProtectedEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := Filter{
			Tenants: []int64{11},
		}
		err := auth.MustHavePermissions(r.Context(),
			auth.READ_DEVICES,
			auth.WRITE_DEVICES)
		if err != nil {
			// unauthorized!
			fmt.Println("must have permissions", err)
			web.HTTPError(w, err)
			return
		}

		if len(filter.Tenants) == 0 {
			// In case the filter is left empty, the desired output is data of all tenants this user has access to
			filter.Tenants, err = auth.GetTenants(r.Context())
			fmt.Println(err)
		}

		if !auth.HasPermissionsFor(r.Context(), filter.Tenants...) {
			// unauthorized!
			fmt.Println("no permis")
			web.HTTPError(w, auth.ErrUnauthorized)
			return
		}

		apiKeyManagerRole := auth.NewRole(
			auth.READ_API_KEYS,
			auth.WRITE_API_KEYS)

		fmt.Println("Role API_KEY_MANAGER:", auth.HasRole(r.Context(), apiKeyManagerRole)) // should be false

		// Authorized!
		// Both authorized to read and write and access
		// to all tenants the user has permissions for or just the ones that were inputted by the user
		fmt.Println("Filter", filter)
		fmt.Println("Authorized!")
		// fmt.Println("> Permissions", grant.permissions)
	}
}

func createToken() (string, error) {
	res, err := http.Get(fmt.Sprintf("http://localhost:4467/.well-known/jwks.json"))
	if err != nil {
		fmt.Print("f err", err)
		panic(fmt.Errorf("failed to fetch jwks: %w", err))
	}
	fmt.Println("f stuff", res.StatusCode)
	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(res.Body).Decode(&jwks); err != nil {
		panic(fmt.Errorf("failed to decode jwks: %w", err))
	}
	keys := jwks.Key("387e4978-078b-4664-afb4-cf9142161610")
	if len(keys) == 0 {
		panic(fmt.Errorf("no keys found for token"))
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256,
		jwt.MapClaims{
			"current_tenant_id": 11,
			"permissions": []string{
				auth.READ_DEVICES.String(),
				auth.READ_API_KEYS.String(),
				auth.WRITE_API_KEYS.String(),
				auth.WRITE_DEVICES.String(),
				"asdsad",
			},
			// tenant:123,
			// device:541
			"user_id": 431,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(key())
	if err != nil {
		fmt.Println("cant create token", err)
		return "", err
	}

	return tokenString, nil
}

func decodeBase64(input string) []byte {
	decoded, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil {
		log.Fatal(err)
	}
	return decoded
}

// Function to decode a base64-encoded integer
func decodeBase64Int(input string) int64 {
	decoded, err := base64.RawURLEncoding.DecodeString(input)
	if err != nil {
		log.Fatal(err)
	}
	return int64(new(big.Int).SetBytes(decoded).Uint64())
}

func key() interface{} {
	privateKeyBytes, err := ioutil.ReadFile("")
	if err != nil {
		panic(err)
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
	if err != nil {
		fmt.Println("Error parsing private key:", err)
		panic(err)
	}
	return privateKey
}
