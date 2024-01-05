package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang-jwt/jwt"
	"sensorbucket.nl/sensorbucket/internal/web"
)

var secretKey = []byte("d0n1 wu3ry, ve3y s4f3")

func main() {
	r := chi.NewRouter()

	// Use middleware to validate JWT
	r.Use(protect)

	t, err := createToken("whatever")
	if err != nil {
		panic(err)
	}
	fmt.Println("TOKEN", t)

	// Define your routes
	r.Get("/stuff", someProtectedEndpoint())

	go http.ListenAndServe(":8086", r)

	fmt.Println("Sleeping for 3s...")
	time.Sleep(time.Second * 3)

	c := http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:8086/stuff", nil)
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
		fmt.Println("protected endpoint")
		tenantId, ok := r.Context().Value("tenant_id").(int64)
		if !ok {
			fmt.Println("no tenant!")
			return
		}
		fmt.Println("Tenant", tenantId)
		permissions, ok := r.Context().Value("permissions").([]string)
		if !ok {
			fmt.Println("no permissions!")
			return
		}
		fmt.Println("permissions", permissions)
		w.Write([]byte("Hello, protected route!"))
	}
}

func protect(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			web.HTTPError(w, ErrAuthHeaderMissing)
			return
		}
		if !strings.Contains(auth, "Bearer ") {
			web.HTTPError(w, ErrAuthHeaderInvalidFormat)
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})
		if err == nil && token.Valid && token.Claims.Valid() == nil {
			if claims, ok := token.Claims.(jwt.MapClaims); ok {
				expired := !claims.VerifyExpiresAt(time.Now().Unix(), true)
				if !expired {
					// JWT itself is validated, pass it to the actual endpoint for further authorization
					next.ServeHTTP(w, r.WithContext(
						context.
							WithValue(
								context.WithValue(r.Context(),
									"tenant_id", currentTenantFromClaims(claims)),
								"permissions",
								permissionsFromClaims(claims),
							),
					))
					return
				}

			}
		}
		web.HTTPError(w, ErrUnauthorized)
	})
}

// TODO: return error instead of empty value
func currentTenantFromClaims(claims jwt.MapClaims) int64 {
	tenant, ok := claims["tenant_id"]
	if ok {
		// TODO: check tenant to string viability
		tenantId, err := strconv.ParseInt(tenant.(string), 10, 32)
		if err == nil {
			return tenantId
		}
	}
	return -1
}

// TODO: return error instead of empty value
func permissionsFromClaims(claims jwt.MapClaims) []string {
	permissions, ok := claims["permissions"]
	if ok {
		if asSlice, ok := permissions.([]string); ok {
			return asSlice
		}
	}
	return []string{}
}

func createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"tenant_id": 42,
			"permissions": []string{
				READ_DEVICES,
				WRITE_DEVICES,
			},
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
