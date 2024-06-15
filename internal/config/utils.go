package config

import (
	"errors"
	"os"
)

// getPrefix returns the prefix for the environment variables (as required by koanf)
func getPrefix() (string, error) {
	prefix := os.Getenv("CONFIG_PREFIX")
	if prefix == "" {
		return "", errors.New("CONFIG_PREFIX is not set")
	}

	return prefix, nil
}
