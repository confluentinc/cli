/*
Package editor allows your CLI users to edit arbitrary data in their preferred editor.

It's just like editing messages in "git commit" or resources with "kubectl edit".

Existing File

The most basic usage is to prompt the user to edit an existing file. This may
be useful to edit the application configuration or a system file, for example.

    edit := editor.NewEditor()
    err := edit.Launch("/etc/bashrc")

Arbitrary Data

Most of the time, the data you want your user to edit isn't in an local file.
In these cases, if you can represent your data in a human editable format
(txt, yaml, hcl, json, etc), then go-editor will enable the user to edit it.

Provide any "io.Reader" with the initial contents:

	original := bytes.NewBufferString("something to be edited\n")

	edit := editor.NewEditor()
	edited, file, err := edit.LaunchTempFile("example", original)
	defer os.Remove(file)
	if err != nil {
	    // handle it
	}

The library leaves it up to you to cleanup the temp file.

This enables your CLI to validate the edited data and prompt the user to
continue editing where they left off, rather than starting over. And if
that's what you want...

Input Validation

If you would like to validate the edited data, use a ValidatingEditor instead.
This will prompt the user to continue editing until validation succeeds or
the edit is cancelled.

Simply create a schema and pass it to the editor:

    schema := &mySchema{}
    edit := editor.NewValidatingEditor(schema)

A schema must implement the Schema interface: https://godoc.org/github.com/confluentinc/go-editor#Schema
*/
package editor
