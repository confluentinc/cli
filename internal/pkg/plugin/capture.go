package plugin

import (
	"bufio"
	"os"
	"os/exec"
)

// Capture the stdout and stderr streams to the command, and interactively pass data in to stdin.
func Capture(command *exec.Cmd, stdin, stdout, stderr chan string) {
	stdinPipe, _ := command.StdinPipe()
	stdoutPipe, _ := command.StdoutPipe()
	stderrPipe, _ := command.StderrPipe()

	go func() {
		for in := range stdin {
			_, _ = stdinPipe.Write([]byte(in))
			stdout <- in
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdoutPipe)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			stdout <- scanner.Text()
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderrPipe)
		scanner.Split(bufio.ScanBytes)
		for scanner.Scan() {
			stderr <- scanner.Text()
		}
	}()

	if err, ok := command.Run().(*exec.ExitError); err != nil && ok {
		os.Exit(err.ExitCode())
	}
}
