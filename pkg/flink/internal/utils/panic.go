package utils

import (
	"time"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type PanicRecoveryWithLimit struct {
	recovers          int
	lastRecover       time.Time
	maxRecovers       int
	timeBetweenPanics time.Duration
}

func NewPanicRecoveryWithLimit(maxRecovers int, timeBetweenPanics time.Duration) PanicRecoveryWithLimit {
	return PanicRecoveryWithLimit{
		recovers:          0,
		maxRecovers:       maxRecovers,
		lastRecover:       time.Unix(0, 0),
		timeBetweenPanics: timeBetweenPanics,
	}
}

func (p *PanicRecoveryWithLimit) WithCustomPanicRecovery(fn func(), customRecovery func()) error {
	WithCustomPanicRecovery(fn, func() {
		if time.Since(p.lastRecover) > p.timeBetweenPanics {
			p.recovers = 0
		}
		p.recovers++
		p.lastRecover = time.Now()
		customRecovery()
	})()

	if p.recovers > p.maxRecovers {
		return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, "Run `confluent flink shell -vvv` to enable debug logs when starting the flink shell and report the output to the CLI team. Kindly share steps reproduce, if possible.\nPlease, restart the CLI.")
	}
	return nil
}

func WithPanicRecovery(fn func()) func() {
	return WithCustomPanicRecovery(fn, func() {
		OutputErr("Error: internal error occurred")
	})
}

func WithCustomPanicRecovery(fn func(), customRecovery func()) func() {
	return func() {
		defer func() {
			if r := recover(); r != nil {
				log.CliLogger.Debug(r)
				if customRecovery != nil {
					WithPanicRecovery(customRecovery)()
				}
			}
		}()
		fn()
	}
}
