package spinner

import (
	"time"

	"github.com/confluentinc/cli/v4/pkg/output"
)

var frames = []string{"|", "/", "-", "\\"}

type Spinner struct {
	stop chan bool
	wait chan bool
}

func New() *Spinner {
	return &Spinner{
		stop: make(chan bool),
		wait: make(chan bool),
	}
}

func (s *Spinner) Start() {
	go s.run()
}

func (s *Spinner) Stop() {
	close(s.stop)
	<-s.wait
}

func (s *Spinner) run() {
	ticker := time.NewTicker(time.Second / 3)
	defer ticker.Stop()

	i := 0

	for {
		select {
		case <-s.stop:
			output.ErrPrint(false, "\033[1D")
			close(s.wait)
			return
		case <-ticker.C:
			output.ErrPrint(false, "\033[1D")
			output.ErrPrint(false, frames[i])
			i = (i + 1) % len(frames)
		}
	}
}
