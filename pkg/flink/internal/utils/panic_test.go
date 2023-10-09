package utils

import (
	"sync"
	"testing"
)

func TestPanicRecoveryWrapper(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)

	go WithPanicRecovery(func() {
		defer wg.Done()
		panic("a panic in a goroutine")
	})()

	wg.Wait()
}
