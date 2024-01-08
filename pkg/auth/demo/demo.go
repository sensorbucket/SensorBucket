package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
	"sensorbucket.nl/sensorbucket/pkg/auth"
)

var secretKey = []byte("d0n1 wu3ry, ve3y s4f3")

func main() {
	r := chi.NewRouter()

	// Use middleware to validate JWT
	r.Use(auth.Protect(secretKey))

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
		grant, err := auth.MustHavePermissions(r.Context(),
			auth.READ_DEVICES,
			auth.WRITE_DEVICES)
		if err != nil {
			// unauthorized!
			web.HTTPError(w, err)
			return
		}

		if len(filter.Tenants) == 0 {
			// In case the filter is left empty, the desired output is data of all tenants this user has access to
			filter.Tenants = grant.GetTenants()
		}

		if !grant.HasPermissionsFor(filter.Tenants...) {
			// unauthorized!
			web.HTTPError(w, auth.ErrUnauthorized)
			return
		}

		fmt.Println("Role API_KEY_MANAGER:", grant.HasRole(auth.API_KEY_MANAGER)) // should be false
		fmt.Println("Role DEVICE_MANAGER:", grant.HasRole(auth.DEVICE_MANAGER))   // should be true

		// Authorized!
		// Both authorized to read and write and access
		// to all tenants the user has permissions for or just the ones that were inputted by the user
		fmt.Println("Filter", filter)
		fmt.Println("Authorized!", grant)
		fmt.Println("> User ID", grant.GetUser())
		fmt.Println("> Tenants", grant.GetTenants())
		// fmt.Println("> Permissions", grant.permissions)
	}
}

func createToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"current_tenant_id": 11,
			"permissions": []string{
				auth.READ_DEVICES.String(),
				auth.READ_API_KEYS.String(),
				auth.WRITE_DEVICES.String(),
				"asdsad",
			},
			"roles": []string{}, // ??
			// tenant:123,
			// device:541
			"user_id": 431,
			"exp":     time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
