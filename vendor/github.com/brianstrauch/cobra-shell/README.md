# cobra-shell

![logo](https://cobra.dev/home/logo.png)

## Description

Leverages the Cobra completion API to generate an interactive shell for any [Cobra](https://github.com/spf13/cobra) CLI, powered by [go-prompt](https://github.com/c-bata/go-prompt).

* On-the-fly autocompletion for all commands
* Static and dynamic autocompletion for args and flags, as described [here](https://github.com/spf13/cobra/blob/master/shell_completions.md)
* Full prompt customizability

## Usage

<img src="https://user-images.githubusercontent.com/7474900/144362981-68bcb984-fbfd-4df5-9194-8d9e92bd65ca.gif" width="50%" />

## Download

```
go get github.com/brianstrauch/cobra-shell
```

## Example

```
package main

import (
    shell "github.com/brianstrauch/cobra-shell"
    "github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{Use: "example"}
	cmd.AddCommand(shell.New())
	_ = cmd.Execute()
}
```
