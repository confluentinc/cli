package config

import (
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/stretchr/testify/require"
)

func signedToken(t *testing.T, claims map[string]any) string {
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.HS256, Key: []byte("0123456789abcdef0123456789abcdef")}, nil)
	require.NoError(t, err)
	token, err := jwt.Signed(signer).Claims(claims).Serialize()
	require.NoError(t, err)
	return token
}

func TestContextState_IsExpired(t *testing.T) {
	for _, tt := range []struct {
		name   string
		claims map[string]any
		want   bool
	}{
		{"valid", map[string]any{"exp": time.Now().Add(time.Hour).Unix()}, false},
		{"expires within the safety margin", map[string]any{"exp": time.Now().Add(30 * time.Second).Unix()}, true},
		{"expired", map[string]any{"exp": time.Now().Add(-time.Hour).Unix()}, true},
		{"no exp claim", map[string]any{"sub": "user"}, false},
		{"non-numeric exp", map[string]any{"exp": "tomorrow"}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			state := &ContextState{AuthToken: signedToken(t, tt.claims)}
			require.Equal(t, tt.want, state.IsExpired())
		})
	}
}

func TestContextState_IsExpired_InvalidToken(t *testing.T) {
	var nilState *ContextState
	require.False(t, nilState.IsExpired())
	require.False(t, (&ContextState{AuthToken: "not-a-jwt"}).IsExpired())
	require.False(t, (&ContextState{}).IsExpired())
}
