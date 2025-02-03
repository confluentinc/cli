package auth

import (
	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConvertToTypeMapString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		expected := convertDateFormat(timestamp)
		actual := "06/16/2021"
		assert.Equal(t, expected, actual)
	})
	t.Run("fail", func(t *testing.T) {
		date := time.Date(2021, time.April, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		expected := convertDateFormat(timestamp)
		actual := "06/16/2021"
		assert.NotEqual(t, expected, actual)
	})
}
