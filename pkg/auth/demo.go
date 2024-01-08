package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var secretKey = []byte("d0n1 wu3ry, ve3y s4f3")

func main() {
	r := chi.NewRouter()

	// Use middleware to validate JWT
	r.Use(Protect(secretKey))

	t, err := createToken("whatever")
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

func someProtectedEndpoint() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := MustHavePermissions(r.Context(), READ_DEVICES, WRITE_DEVICES); err != nil {
			// Not authorized!
			web.HTTPError(w, err)
			return
		}

		// Authorized!
		w.Write([]byte("Hello, protected route!"))
	}
}

func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"current_tenant_id": 42,
			"permissions": []string{
				READ_DEVICES,
				WRITE_DEVICES,
			},
			"username": username,
			"exp":      time.Now().Add(-time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
