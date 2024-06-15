package shallenge

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log/slog"
	"math"
	"math/big"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelmetric "go.opentelemetry.io/otel/metric"
)

var (
	meter = otel.Meter("github.com/michaelpeterswa/shallenge-miner/internal/shallenge") // meter is the named meter for the spider package

	hashesMadeMetric otelmetric.Int64Counter // hashMadeMetric is the metric for the number of hashes made
)

func init() {
	var err error

	hashesMadeMetric, err = meter.Int64Counter("hashes_made")
	if err != nil {
		slog.Error("failed to create hashesMade metric", slog.String("err", err.Error()))
		os.Exit(1)
	}
}

func otelHashQualityAttribute(quality float64) otelmetric.MeasurementOption {
	return otelmetric.WithAttributes(attribute.Int("quality", ditherQuality(quality)))
}

func ditherQuality(quality float64) int {
	intQuality := int(math.Ceil(quality*100/5)) * 5

	return intQuality
}

type HashQuality struct {
	Name    string
	Nonce   string
	Sha256  string
	Quality float64
}

func NonceBuilder(ctx context.Context, i int64) (string, error) {
	bigInt := big.NewInt(i)
	suffix := base64.RawURLEncoding.EncodeToString(bigInt.Bytes())

	if len(suffix) > 32 {
		return "", fmt.Errorf("nonce too long: %d", len(suffix))
	}

	return "N0GpU/N0Pr0bLeM//N0tFundedByYC//" + suffix, nil
}

func HashWithQuality(ctx context.Context, name string, nonce string) *HashQuality {
	input := fmt.Sprintf("%s/%s", name, nonce)

	sha256bytes := sha256.Sum256([]byte(input))
	sha256 := hex.EncodeToString(sha256bytes[:])

	zeroPrefix := 0
	for _, char := range sha256 {
		if char == '0' {
			zeroPrefix++
		} else {
			break
		}
	}

	quality := float64(zeroPrefix) / float64(len(sha256))

	hashesMadeMetric.Add(ctx, 1, otelHashQualityAttribute(quality))

	return &HashQuality{
		Name:    name,
		Nonce:   nonce,
		Sha256:  sha256,
		Quality: quality,
	}
}
