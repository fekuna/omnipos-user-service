package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

func getEnv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	return v
}

func getEnvRequired(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("missing required env: %s", key))
	}
	return v
}

func getBoolEnv(key string, def bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	b, err := strconv.ParseBool(v)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: must be true/false", key))
	}

	return b
}

func getEnvInt(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	val, err := strconv.Atoi(v)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: must be integer", key))
	}

	return val
}

func getEnvDuration(key string, def time.Duration) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return def
	}

	val, err := time.ParseDuration(v)
	if err != nil {
		panic(fmt.Sprintf("invalid %s: must be duration", key))
	}

	return val
}
