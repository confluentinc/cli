package billing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrice(t *testing.T) {
	assert.Equal(t, "$1.00 USD/GB", FormatPrice(1, "GB"))
}

func TestFormatDollars(t *testing.T) {
	assert.Equal(t, "$1.00", FormatDollars(1))
	assert.Equal(t, "$1.20", FormatDollars(1.2))
	assert.Equal(t, "$1.23", FormatDollars(1.23))
	assert.Equal(t, "$1.23456", FormatDollars(1.23456))
	assert.Equal(t, "$0.0041", FormatDollars(0.004100))
}
