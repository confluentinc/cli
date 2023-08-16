package retry

import (
	"fmt"
	"time"

	"github.com/confluentinc/cli/v3/pkg/log"
)

func Retry(tick, timeout time.Duration, f func() error) error {
	if err := f(); err != nil {
		log.CliLogger.Debugf("Fail #1: %v", err)
	} else {
		return nil
	}

	ticker := time.NewTicker(tick)
	after := time.After(timeout)

	for i := 2; true; i++ {
		select {
		case <-ticker.C:
			if err := f(); err != nil {
				log.CliLogger.Debugf("Fail #%d: %v", i, err)
			} else {
				return nil
			}
		case <-after:
			return fmt.Errorf("retry failed due to timeout of %v", timeout)
		}
	}

	return nil
}
