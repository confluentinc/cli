package editor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

// ErrEditing represents an editing error
type ErrEditing error

var (
	msgValidationFailed        = "The edited file failed validation"
	msgCancelledNoValidChanges = "Edit cancelled, no valid changes were saved."
	msgCancelledNoOrigChanges  = "Edit cancelled, no changes made."
	msgCancelledEmptyFile      = "Edit cancelled, saved file was empty."
	msgPreserveFileLocation    = "A copy of your changes has been stored to %s\n"

	defaultInvalidFn = func(e error) error {
		fmt.Printf("%s: %v\n", msgValidationFailed, e)
		return ErrEditing(fmt.Errorf(msgCancelledNoValidChanges))
	}
	defaultNoChangesFn    = func() (bool, error) { return true, ErrEditing(fmt.Errorf(msgCancelledNoOrigChanges)) }
	defaultEmptyFileFn    = func() (bool, error) { return true, ErrEditing(fmt.Errorf(msgCancelledEmptyFile)) }
	defaultPreserveFileFn = func(data []byte, file string, err error) ([]byte, string, error) {
		fmt.Printf(msgPreserveFileLocation, file)
		return data, file, err
	}
	defaultCommentChars = []string{"#", "//"}
)

// ValidatingEditor is an Editor which validates data against a schema. This will
// prompt the user to continue editing until validation succeeds or the edit is cancelled.
type ValidatingEditor struct {
	*BasicEditor

	// Schema is used to validate the edited data.
	Schema Schema

	// InvalidFn is called when a Schema fails to validate data.
	InvalidFn ValidationFailedFn
	// OriginalUnchangedFn is called when no changes were made from the original data.
	OriginalUnchangedFn CancelEditingFn
	// EmptyFileFn is called when the edited data is (effectively) empty; the file doesn't have any uncommented lines (ignoring whitespace)
	EmptyFileFn CancelEditingFn
	// PreserveFileFn is called when a non-recoverable error has occurred and the users edits have been preserved in a temp file.
	PreserveFileFn PreserveFileFn

	// CommentChars is a list of comment string prefixes for determining "empty" files. Defaults to "#" and "//".
	CommentChars []string
}

// NewValidatingEditor returns a new ValidatingEditor.
//
// This extends the BasicEditor with schema validation capabilities.
func NewValidatingEditor(schema Schema) *ValidatingEditor {
	return &ValidatingEditor{
		BasicEditor:         NewEditor(),
		Schema:              schema,
		InvalidFn:           defaultInvalidFn,
		OriginalUnchangedFn: defaultNoChangesFn,
		EmptyFileFn:         defaultEmptyFileFn,
		PreserveFileFn:      defaultPreserveFileFn,
		CommentChars:        defaultCommentChars,
	}
}

// LaunchTempFile launches the users preferred editor on a temporary file.
// This file is initialized with contents from the provided stream and named
// with the given prefix.
//
// Returns the modified data, the path to the temporary file so the caller can
// clean it up, and an error.
//
// A file may be present even when an error is returned. Please clean it up.
//
// The last byte of "obj" must be a newline to cancel editing if no changes are made.
// (This is because many editors like vim automatically add a newline when saving.)
func (e *ValidatingEditor) LaunchTempFile(prefix string, obj io.Reader) ([]byte, string, error) {
	editor := e.BasicEditor.clone()

	var (
		prevErr  error
		original []byte
		edited   []byte
		file     string
		err      error
	)

	originalObj, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", err
	}

	// loop until we succeed or cancel editing
	for {
		// Create the file to edit
		buf := &bytes.Buffer{}
		if prevErr == nil {
			buf.Write(originalObj)
			original = buf.Bytes()
		} else {
			// Preserve the edited file
			buf.Write(edited)
		}

		// Launch the editor
		editedDiff := edited
		edited, file, err = editor.LaunchTempFile(prefix, buf)
		if err != nil {
			return e.PreserveFileFn(edited, file, err)
		}

		// If we're retrying the loop because of an error, and no change was made in the file, short-circuit
		if prevErr != nil && bytes.Equal(editedDiff, edited) {
			return e.PreserveFileFn(edited, file, e.InvalidFn(prevErr))
		}

		// Compare contents for changes
		if bytes.Equal(original, edited) {
			cancel, err := e.OriginalUnchangedFn()
			if cancel {
				os.Remove(file)
				return nil, "", err
			}
		}

		// Check for an (effectively) empty file
		empty, err := e.isEmpty(edited)
		if err != nil {
			return e.PreserveFileFn(edited, file, err)
		}
		if empty {
			cancel, err := e.EmptyFileFn()
			if cancel {
				os.Remove(file)
				return nil, "", err
			}
		}

		// Apply validation
		err = e.Schema.ValidateBytes(edited)
		if err != nil {
			prevErr = err
			os.Remove(file)
			continue
		}

		return edited, file, nil
	}
}

// isEmpty returns true if the file doesn't have any uncommented lines (ignoring whitespace)
func (e *ValidatingEditor) isEmpty(data []byte) (bool, error) {
	empty := true
	scanner := bufio.NewScanner(bytes.NewBuffer(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			commented := false
			for _, c := range e.CommentChars {
				if strings.HasPrefix(line, c) {
					commented = true
				}
			}
			if !commented {
				empty = false
				break
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}
	return empty, nil
}
