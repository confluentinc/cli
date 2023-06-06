package billing

import (
	"fmt"
	"strings"
)

func FormatPrice(price float64, unit string) string {
	priceStr := FormatDollars(price)
	return fmt.Sprintf("%s USD/%s", priceStr, unit)
}

func FormatDollars(amount float64) string {
	amountStr := fmt.Sprint(amount)

	// Require >= 2 digits after the decimal
	if strings.Contains(amountStr, ".") {
		// Extend the remainder if needed
		r := strings.Split(amountStr, ".")
		for len(r[1]) < 2 {
			r[1] += "0"
		}
		amountStr = strings.Join(r, ".")
	} else {
		amountStr += ".00"
	}

	return fmt.Sprintf("$%s", amountStr)
}
