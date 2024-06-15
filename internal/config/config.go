package config

import (
	"fmt"
	"strings"

	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/v2"
)

const (
	LogLevel = "log.level"

	MetricsEnabled = "metrics.enabled"
	MetricsPort    = "metrics.port"

	TracingEnabled    = "tracing.enabled"
	TracingSampleRate = "tracing.samplerate"
	TracingService    = "tracing.service"
	TracingVersion    = "tracing.version"
)

func NewConfig() (*koanf.Koanf, error) {
	prefix, err := getPrefix()
	if err != nil {
		return nil, fmt.Errorf("could not get environment variable prefix: %w", err)
	}

	k := koanf.New(".")

	err = k.Load(env.Provider(prefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, prefix)), "_", ".", -1)
	}), nil)
	if err != nil {
		return nil, fmt.Errorf("could not load environment variables: %w", err)
	}

	return k, nil
}
