package billing

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatBalance(t *testing.T) {
	require.Equal(t, "$0.00/1.00 USD", formatBalance(0, 10000))
}

func TestFormatExpiration(t *testing.T) {
	date := time.Date(2021, time.June, 16, 0, 0, 0, 0, time.Local)
	require.Equal(t, "Jun 16, 2021", formatExpiration(date.Unix()))
}
