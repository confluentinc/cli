package config

import (
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
)

func signedToken(t *testing.T, claims map[string]any) string {
	t.Helper()
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte("0123456789abcdef0123456789abcdef")}, nil)
	require.NoError(t, err)
	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(t, err)
	return token
}

func TestContextState_IsExpired(t *testing.T) {
	now := time.Unix(1_700_000_000, 0)

	tests := []struct {
		name  string
		state func(t *testing.T) *ContextState
		want  bool
	}{
		{"nil state", func(*testing.T) *ContextState { return nil }, false},
		{"valid, well before margin", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(time.Hour).Unix()})}
		}, false},
		{"valid, 61s left", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(61 * time.Second).Unix()})}
		}, false},
		{"exactly at margin (60s left)", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(60 * time.Second).Unix()})}
		}, true},
		{"30s left", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(30 * time.Second).Unix()})}
		}, true},
		{"just expired", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(-time.Second).Unix()})}
		}, true},
		{"long expired", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": now.Add(-24 * time.Hour).Unix()})}
		}, true},
		{"malformed token", func(*testing.T) *ContextState { return &ContextState{AuthToken: "not-a-jwt"} }, false},
		{"empty token", func(*testing.T) *ContextState { return &ContextState{} }, false},
		{"no exp claim", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"sub": "user"})}
		}, false},
		{"non-numeric exp", func(t *testing.T) *ContextState {
			return &ContextState{AuthToken: signedToken(t, map[string]any{"exp": "tomorrow"})}
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.state(t).isExpiredAt(now))
		})
	}
}
