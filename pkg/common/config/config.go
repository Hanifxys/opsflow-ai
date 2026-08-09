// Package config provides environment-based configuration loading.
// No external dependencies — uses os.Getenv with typed helpers.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// GetEnv returns the value of the environment variable or the fallback.
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// MustGetEnv returns the value of the environment variable or panics.
func MustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %s is not set", key))
	}
	return v
}

// GetEnvInt returns the environment variable as int or the fallback.
func GetEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return i
}

// GetEnvBool returns the environment variable as bool or the fallback.
func GetEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}
