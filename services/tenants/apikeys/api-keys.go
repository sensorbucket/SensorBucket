package apikeys

import (
	"math/rand"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type TenantState string

var (
	Active TenantState = "Active"
)

type Tenant struct {
	ID    int64
	Name  string
	State TenantState
}

func newApiKey(name string, expirationDate *time.Time) ApiKey {
	return ApiKey{
		Key: Key{
			ID:             rand.Int63(),
			Name:           name,
			ExpirationDate: expirationDate,
		},
		Value: generateRandomString32(),
	}
}

type ApiKey struct {
	Key
	Value string
}
type HashedApiKey struct {
	Key
	Value    string
	TenantID int64
}

type Key struct {
	ID             int64
	Name           string
	ExpirationDate *time.Time
}

func (k *Key) IsExpired() bool {
	if k.ExpirationDate == nil {
		// Expiration date is optional
		return false
	}
	return k.ExpirationDate.Before(time.Now())
}

func (a *ApiKey) hash() (HashedApiKey, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(a.Value), bcrypt.DefaultCost)
	if err != nil {
		return HashedApiKey{}, err
	}
	return HashedApiKey{
		Key:   a.Key,
		Value: string(b),
	}, nil
}

func (a *HashedApiKey) compare(with string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.Value), []byte(with)) == nil
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
