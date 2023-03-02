package spinner

import (
	"time"

	"github.com/confluentinc/cli/internal/pkg/output"
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
			clear()
			close(s.wait)
			return
		case <-ticker.C:
			clear()
			output.ErrPrint(frames[i])
			i = (i + 1) % len(frames)
		}
	}
}

func clear() {
	output.ErrPrint("\033[1D")
}
