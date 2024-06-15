package main

import (
	"context"
	"log/slog"
	"math/rand"
	"os"
	"time"

	"github.com/alpineworks/ootel"
	"github.com/michaelpeterswa/shallenge-miner/internal/config"
	"github.com/michaelpeterswa/shallenge-miner/internal/db"
	"github.com/michaelpeterswa/shallenge-miner/internal/dragonfly"
	"github.com/michaelpeterswa/shallenge-miner/internal/logging"
	"github.com/michaelpeterswa/shallenge-miner/internal/shallenge"
	"github.com/sourcegraph/conc/pool"
)

func main() {
	ctx := context.Background()

	slogHandler := slog.NewJSONHandler(os.Stdout, nil)
	slog.SetDefault(slog.New(slogHandler))

	slog.Info("welcome to shallenge-miner!")

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

	dbClient, err := db.NewDBClient(ctx, c.String(config.PostgresConn))
	if err != nil {
		slog.Error("could not create db client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbClient.Close()

	dfClient, err := dragonfly.NewDragonflyClient(ctx, c.String(config.DragonflyHost), c.Int(config.DragonflyPort))
	if err != nil {
		slog.Error("could not create dragonfly client", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dfClient.Close()

	inChan := make(chan int64, 1000000)

	writerPool := pool.New()
	for i := 0; i < c.Int(config.WriterWorkers); i++ {
		writerPool.Go(func() {
			for {
				time.Sleep(time.Duration(rand.Intn(int(2*time.Second))) + c.Duration(config.BatchDelay))
				min, max, err := dfClient.UpsertBatchCounter(ctx, "batchmax", int64(c.Int(config.BatchSize)))
				if err != nil {
					slog.Error("could not upsert batch counter", slog.String("error", err.Error()))
					continue
				}

				for i := min; i < max; i++ {
					inChan <- i
					i++
				}
			}
		})
	}

	readerPool := pool.New()
	for i := 0; i < c.Int(config.ReaderWorkers); i++ {
		writerPool.Go(func() {
			for i := range inChan {
				nonce, err := shallenge.NonceBuilder(ctx, int64(i))
				if err != nil {
					slog.Error("could not build nonce", slog.String("error", err.Error()))
					continue
				}

				hwq := shallenge.HashWithQuality(ctx, "nwradio", nonce)

				if hwq.Quality >= 0.05 {
					err = dbClient.InsertResult(ctx, hwq.Name, hwq.Nonce, hwq.Sha256, hwq.Quality)
					if err != nil {
						slog.Error("could not insert result", slog.String("error", err.Error()))
						continue
					}
				}
			}
		})
	}

	writerPool.Wait()
	readerPool.Wait()
}
