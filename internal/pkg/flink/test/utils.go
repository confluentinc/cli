package test

import (
	"os"

	"github.com/stretchr/testify/require"
)

func RunAndCaptureSTDOUT(t require.TestingT, test func()) string {
	// Redirect STDOUT to a buffer
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Run the test
	test()

	// Close the writer and restore the original STDOUT
	err := w.Close()
	require.NoError(t, err)
	os.Stdout = old

	// Read the output from the buffer
	output := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output <- string(buf[:n])
	}()
	return <-output
}
