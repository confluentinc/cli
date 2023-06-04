package config

import "time"

const (
	InitialWaitTime        = time.Millisecond * 300
	WaitTimeIncrease       = 300
	DefaultTimeoutDuration = 10 * time.Minute
)
