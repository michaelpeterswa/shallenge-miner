package shallenge_test

import (
	"context"
	"testing"

	"github.com/michaelpeterswa/shallenge-miner/internal/shallenge"
	"github.com/stretchr/testify/assert"
)

func TestHashWithQuality(t *testing.T) {
	tests := []struct {
		testName string
		name     string
		nonce    string
		expected *shallenge.HashQuality
	}{
		{
			testName: "basic test",
			name:     "asdf1234",
			nonce:    "1234",
			expected: &shallenge.HashQuality{
				Name:    "asdf1234",
				Nonce:   "1234",
				Sha256:  "170fab442ecea6639a617ab74e61805e282eca528ac1b16273e6dfd5998d55fb",
				Quality: 0,
			},
		},
		{
			testName: "basic test with quality",
			name:     "jonas-w",
			nonce:    "00000000000000000000000001G3mLyq",
			expected: &shallenge.HashQuality{
				Name:    "jonas-w",
				Nonce:   "00000000000000000000000001G3mLyq",
				Sha256:  "00000000000245db756318cbfae7f1b874b680cd74fdc73e61c292cba91e7b3f",
				Quality: 0.171875,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			ctx := context.Background()
			assert.Equal(t, tc.expected, shallenge.HashWithQuality(ctx, tc.name, tc.nonce))
		})
	}
}
