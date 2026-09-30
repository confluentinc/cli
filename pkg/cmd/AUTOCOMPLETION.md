# Autocompletion

We leverage the [Cobra Completions API](https://github.com/spf13/cobra/blob/master/shell_completions.md) to provide tab
autocompletion for all of our commands, arguments, and flags. Autocompletion is used extensively throughout the CLI, and
occasionally requires authentication, so we provide several helper functions to limit the amount of boilerplate.

Here's how you would use the `NewValidArgsFunction()` function to dynamically complete a command which accepts an API
key as an argument:

```go
pcmd.NewValidArgsFunction(c.validArgs)
```

```go
func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
    if len(args) > 0 {
        return nil
    }

    if err := c.PersistentPreRunE(cmd, args); err != nil {
        return nil
    }

    return pcmd.AutocompleteApiKeys(c.V2Client)
}
```

Here's how you would use the `RegisterFlagCompletionFunc()` function to dynamically complete a flag which accepts an API
key as a value:

```go
pcmd.RegisterFlagCompletionFunc(cmd, "api-key", func(cmd *cobra.Command, args []string) []string {
    if err := c.PersistentPreRunE(cmd, args); err != nil {
        return nil
    }

    return pcmd.AutocompleteApiKeys(c.V2Client)
})
```