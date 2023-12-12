package utils

import (
	"time"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type PanicRecovererWithLimitImpl struct {
	recovers          int
	lastRecover       time.Time
	maxRecovers       int
	timeBetweenPanics time.Duration
}

type PanicRecovererWithLimit interface {
	WithCustomPanicRecovery(fn func(), customRecovery func()) func() error
}

func NewPanicRecovererWithLimit(maxRecovers int, timeBetweenPanics time.Duration) PanicRecovererWithLimit {
	recoverer := PanicRecovererWithLimitImpl{
		recovers:          0,
		maxRecovers:       maxRecovers,
		lastRecover:       time.Unix(0, 0),
		timeBetweenPanics: timeBetweenPanics,
	}
	return &recoverer
}

func (p *PanicRecovererWithLimitImpl) WithCustomPanicRecovery(fn func(), customRecovery func()) func() error {
	if time.Since(p.lastRecover) > p.timeBetweenPanics {
		p.recovers = 0
	}

	p.recovers++
	if p.recovers > p.maxRecovers {
		return func() error {
			return errors.NewErrorWithSuggestions(errors.InternalServerErrorMsg, "Run `confluent flink shell -vvv` to enable debug logs when starting the flink shell and report the output to the CLI team. Kindly share steps reproduce, if possible.\nPlease, restart the CLI.")
		}
	}

	p.lastRecover = time.Now()
	return func() error {
		WithCustomPanicRecovery(fn, customRecovery)()
		return nil
	}
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
