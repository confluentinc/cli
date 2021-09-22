package spinner

import (
	"fmt"
	"os"
	"time"
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
			fmt.Fprint(os.Stderr, frames[i])
			i = (i + 1) % len(frames)
		}
	}
}

func clear() {
	fmt.Fprint(os.Stderr, "\033[1D")
}
