# go-ps1

## Description

This package provides a [Cobra](https://github.com/spf13/cobra) command that helps your users add custom PS1 prompts for your CLI.
Users can customize their own PS1 prompt using syntax from the Go [template](https://pkg.go.dev/text/template) package and formatting tokens.

## Example

1. Initialize a PS1 object with a list of tokens that users can use to customize their prompt (for instance, the %C token below will display the user's login context) and add the `prompt` command to your Cobra CLI.
```go
package main

import (
	"github.com/confluentinc/go-ps1"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{Use: "confluent"}
	
	ctx := context.Load()
	prompt := ps1.New(root.Name(), []ps1.Token{
		{
			Name: 'C',
			Desc: "Print the current login context.",
			Func: func() string { return ctx.Name() },
		},
	})

	root.AddCommand(prompt.NewCommand())
}
```

2. Use this command to print a custom string consisting of the CLI name and the user's login context.
```bash
$ confluent prompt -f '(confluent|%C)'
(confluent|context-1)
```

3. Then, append it to the `$PS1` environment variable to continuously display it in the terminal prompt. (Note: ZSH users will have to run `setopt prompt_subst` prior to this step.)
```bash
$ export PS1='$(confluent prompt -f "(confluent|%C)") '$PS1
(confluent|context-1) $ confluent context use context-2
(confluent|context-2) $ ...
```

## Additional Features

Colors and other styling attributes can be added to the format string. To avoid escaping special characters, consider defining this string in a separate variable.
```bash
$ format='({{fgcolor "blue" "confluent" | attr "bold"}}|{{fgcolor "red" %C}})'
$ export PS1='$(confluent prompt -f $format) '$PS1
```
