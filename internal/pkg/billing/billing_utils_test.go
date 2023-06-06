package billing

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFormatPrice(t *testing.T) {
	assert.Equal(t, "$1.00 USD/GB", FormatPrice(1, "GB"))
	assert.Equal(t, "$1.20 USD/GB", FormatPrice(1.2, "GB"))
	assert.Equal(t, "$1.23 USD/GB", FormatPrice(1.23, "GB"))
	assert.Equal(t, "$1.23456 USD/GB", FormatPrice(1.23456, "GB"))
}
