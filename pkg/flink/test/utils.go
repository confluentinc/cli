package test

import (
	"os"

	"github.com/stretchr/testify/require"
)

func RunAndCaptureSTDOUT(t require.TestingT, test func()) string {
	// Redirect STDOUT to a buffer
	old := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	defer func() {
		// Close the writer and restore the original STDOUT
		os.Stdout = old
		require.NoError(t, err)
	}()

	// Run the test
	test()

	err = w.Close()

	// Read the output from the buffer
	output := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		output <- string(buf[:n])
	}()
	return <-output
}
