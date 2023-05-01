package price

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatPrice(t *testing.T) {
	assert.Equal(t, "$1.00 USD/GB", formatPrice(1, "GB"))
	assert.Equal(t, "$1.20 USD/GB", formatPrice(1.2, "GB"))
	assert.Equal(t, "$1.23 USD/GB", formatPrice(1.23, "GB"))
	assert.Equal(t, "$1.234 USD/GB", formatPrice(1.234, "GB"))
}
