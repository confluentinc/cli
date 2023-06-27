package config

import "time"

const (
	InitialWaitTime         = 300 * time.Millisecond
	WaitTimeIncrease        = 300
	DefaultTimeoutDuration  = 10 * time.Minute
	ShouldCleanupStatements = false
)
