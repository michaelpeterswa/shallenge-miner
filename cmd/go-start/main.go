package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/alpineworks/ootel"
	"github.com/michaelpeterswa/go-start/internal/config"
	"github.com/michaelpeterswa/go-start/internal/logging"
)

func main() {
	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(slogHandler))

	slog.Info("welcome to go-start!")

	c, err := config.NewConfig()
	if err != nil {
		slog.Error("could not create config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slogLevel, err := logging.LogLevelToSlogLevel(c.String(config.LogLevel))
	if err != nil {
		slog.Error("could not parse log level", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.SetLogLoggerLevel(slogLevel)

	ctx := context.Background()

	ootelClient := ootel.NewOotelClient(
		ootel.WithMetricConfig(
			ootel.NewMetricConfig(
				c.Bool(config.MetricsEnabled),
				c.Int(config.MetricsPort),
			),
		),
		ootel.WithTraceConfig(
			ootel.NewTraceConfig(
				c.Bool(config.TracingEnabled),
				c.Float64(config.TracingSampleRate),
				c.String(config.TracingService),
				c.String(config.TracingVersion),
			),
		),
	)

	shutdown, err := ootelClient.Init(ctx)
	if err != nil {
		panic(err)
	}

	defer func() {
		_ = shutdown(ctx)
	}()

	<-time.After(2 * time.Minute)
}
