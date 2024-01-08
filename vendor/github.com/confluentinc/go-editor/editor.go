package editor

import (
	"io"
	"os"
	"os/exec"
	"runtime"
)

var (
	editor = "vim"
)

func init() {
	if runtime.GOOS == "windows" {
		editor = "notepad"
	}
	if e := os.Getenv("VISUAL"); e != "" {
		editor = e
	} else if e := os.Getenv("EDITOR"); e != "" {
		editor = e
	}
}

// BasicEditor launches an editor given by a specific command.
type BasicEditor struct {
	Command string
	// this is only for testing
	LaunchFn func(command, file string) error
}

// NewEditor launches an instance of the users preferred editor. The editor
// to use is determined by reading the $VISUAL and $EDITOR environment variables.
// If neither of these are present, vim or notepad (on Windows) is used.
func NewEditor() *BasicEditor {
	return &BasicEditor{
		Command:  editor,
		LaunchFn: launch,
	}
}

func (e *BasicEditor) clone() *BasicEditor {
	return &BasicEditor{
		Command:  e.Command,
		LaunchFn: e.LaunchFn,
	}
}

// Launch opens the given file path in the external editor or returns an error.
func (e *BasicEditor) Launch(file string) error {
	return e.LaunchFn(e.Command, file)
}

func launch(command, file string) error {
	cmd := exec.Command(command, file)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// LaunchTempFile launches the users preferred editor on a temporary file.
// This file is initialized with contents from the provided stream and named
// with the given prefix.
//
// Returns the modified data, the path to the temporary file so the caller can
// clean it up, and an error.
//
// A file may be present even when an error is returned. Please clean it up.
func (e *BasicEditor) LaunchTempFile(prefix string, r io.Reader) ([]byte, string, error) {
	f, err := os.CreateTemp("", prefix)
	if err != nil {
		return nil, "", err
	}
	defer f.Close()

	// seed the editor with the initial temp file contents
	if _, err := io.Copy(f, r); err != nil {
		os.Remove(f.Name())
		return nil, "", err
	}

	// close the fd to prevent the editor being unable to save file
	if err := f.Close(); err != nil {
		return nil, "", err
	}

	// launch the external editor on the temp file
	if err := e.Launch(f.Name()); err != nil {
		return nil, f.Name(), err
	}

	bytes, err := os.ReadFile(f.Name())
	return bytes, f.Name(), err
}
