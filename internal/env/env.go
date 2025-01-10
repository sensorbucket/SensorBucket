package env

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

func Could(key, value string) string {
	v := os.Getenv(key)
	if v == "" {
		v = value
	}
	return v
}

func CouldInt(key string, fallback int) int {
	str := os.Getenv(key)
	if str == "" {
		return fallback
	}
	v, err := strconv.Atoi(str)
	if err != nil {
		log.Fatalf("env %s should be an integer: %s\n", key, err.Error())
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
