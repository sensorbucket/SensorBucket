package apikeys

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ApiKey struct {
	Key
	Secret string
}
type HashedApiKey struct {
	Key
	SecretHash string
	TenantID   int64
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
	b, err := bcrypt.GenerateFromPassword([]byte(a.Secret), bcrypt.DefaultCost)
	if err != nil {
		return HashedApiKey{}, err
	}
	return HashedApiKey{
		Key:        a.Key,
		SecretHash: string(b),
	}, nil
}

func (a *HashedApiKey) compare(with string) bool {
	return bcrypt.CompareHashAndPassword([]byte(a.SecretHash), []byte(with)) == nil
}

func newApiKey(name string, expirationDate *time.Time) (ApiKey, error) {
	secret, err := generateRandomString()
	if err != nil {
		return ApiKey{}, err
	}
	id, err := generateRandomInt64()
	if err != nil {
		return ApiKey{}, err
	}
	return ApiKey{
		Key: Key{
			ID:             id,
			Name:           name,
			ExpirationDate: expirationDate,
		},
		Secret: secret,
	}, nil
}

func generateRandomString() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", nil
	}
	res := hex.EncodeToString(b)
	return res, nil
}

func generateRandomInt64() (int64, error) {
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return 0, err
	}
	res := int64(binary.BigEndian.Uint64(randomBytes[:8]))
	if res < 0 {
		res = -res
	}
	return int64(res), nil
}
