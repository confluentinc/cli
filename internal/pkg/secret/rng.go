package secret

import (
	crand "crypto/rand"
	"encoding/binary"
	mrand "math/rand"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Provide an entropy-based source of randomness.
// This allows us to use math/rand for cryptographically-secure operations.
//
// Using math/rand instead of crypt/rand directly also lets us inject a
// pseudo-random (seeded) source for testing.
//
// Based on https://blog.gopheracademy.com/advent-2017/a-tale-of-two-rands/
// https://github.com/orion-labs/go-crypto-source - Apache 2.0
type cryptoSource struct {
	Logger *log.Logger
}

var _ mrand.Source = (*cryptoSource)(nil)

func (s *cryptoSource) Seed(_ int64) {
	// no-op - seeds are only for pseudo-random RNGs,
	// not cryptographically-secure / entropy-driven RNGs
}

func (s *cryptoSource) Int63() int64 {
	return int64(s.Uint64() & ^uint64(1<<63))
}

func (s *cryptoSource) Uint64() (v uint64) {
	err := binary.Read(crand.Reader, binary.BigEndian, &v)
	if err != nil {
		s.Logger.Error(err)
	}
	return v
}
