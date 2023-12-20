package tenants

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type Tenant struct {
	ID   int64
	Name string
}

func NewApiKey() ApiKey {
	return ApiKey{
		ID:    rand.Int63(),
		Value: generateRandomString32(),
	}
}

type ApiKey struct {
	ID          int64
	HashedValue string
	Value       string
}

func (a *ApiKey) Hash() error {
	b, err := bcrypt.GenerateFromPassword([]byte(a.Value), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	a.HashedValue = string(b)
	return nil
}

func (a *ApiKey) Compare(with string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.HashedValue), []byte(with)) == nil
}

func generateRandomString32() string {
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	randomString := make([]byte, 32)
	for i := range randomString {
		randomString[i] = charset[rand.Intn(len(charset))]
	}

	return string(randomString)
}
