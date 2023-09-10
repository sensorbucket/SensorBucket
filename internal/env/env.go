package env

import (
	"fmt"
	"os"
)

func Could(key, value string) string {
	v := os.Getenv(key)
	if v == "" {
		v = value
	}
	return v
}

func Must(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("environment variable (%s) must be set", key))
	}
	return v
}
