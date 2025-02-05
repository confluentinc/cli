package auth

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
)

func TestConvertDateFormatToString(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		actual := convertDateFormat(timestamp)
		expected := "06/16/2021"
		assert.Equal(t, expected, actual)
	})
	t.Run("fail", func(t *testing.T) {
		date := time.Date(2021, time.April, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		actual := convertDateFormat(timestamp)
		expected := "06/16/2021"
		assert.NotEqual(t, expected, actual)
	})
	t.Run("fail, nil", func(t *testing.T) {
		actual := convertDateFormat(&types.Timestamp{})
		expected := "Invalid Date"
		assert.Equal(t, expected, actual)
	})
	t.Run("fail, outside range", func(t *testing.T) {
		date := time.Date(2199, time.April, 16, 12, 0, 0, 0, time.UTC)
		timestamp := &types.Timestamp{Seconds: date.Unix()}
		actual := convertDateFormat(timestamp)
		expected := "Invalid Date"
		assert.Equal(t, expected, actual)
	})
}
