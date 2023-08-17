package flink

import (
	"fmt"
	"time"
)

// GetLocalTimezone returns the local timezone in a custom format along with the UTC offset
// Example: UTC+02:00 or UTC-08:00
func GetLocalTimezone() string {
	_, offsetSeconds := time.Now().Zone()
	return formatUtcOffsetToTimezone(offsetSeconds)
}

func formatUtcOffsetToTimezone(offsetSeconds int) string {
	timeOffset := time.Duration(offsetSeconds) * time.Second
	sign := "+"
	if offsetSeconds < 0 {
		sign = "-"
		timeOffset *= -1
	}
	offsetStr := fmt.Sprintf("%02d:%02d", int(timeOffset.Hours()), int(timeOffset.Minutes())%60)
	return fmt.Sprintf("UTC%s%s", sign, offsetStr)
}
