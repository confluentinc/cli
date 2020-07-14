# Errors And Messages Handling

The CLI codebase stores the strings of all messages in one location: the errors package [internal/pkg/errors](internal/pkg/errors). This encompasses all kinds of communication with the users, whether it be error messages, or message indicating success (e.g. successfully deleting a Kafka cluster). The goal is to ease the verification of consistency and correctness of messages, and to simplify future work in internationalizing the CLI.

## General Message Format
- `""` surrounding names and ID’s
    - e.g. `Check that the resource "lkc-abc" exists.`, `Error: topic "bob" does not exist`
- ```` `` ```` surrounding commands or flags
    - e.g. ``You must pass `--cluster` flag with the command or set an active Kafka in your context with `ccloud kafka cluster use`.``

## Creating an Error
### Message Format
Error message format
- short one line description describing what the error is
- do not capitalize the first letter (unless the first word is a name)
- do not add full stop at the end
- the variable name must end with *ErrorMsg*

Any suggestions of the type of actions the users should take, or any longer description should be kept in the suggestions section.
Suggestion message is not required.
Suggestions format
- full sentence
- capitalize first letter
- end with a full stop 
- the variable name must end with *Suggestions*

examples:
```
> ./ccloud ksql app create kk --cluster lkc-asfdsaf
Error: Kafka cluster "lkc-asfdsaf" not found

Suggestions:
    List Kafka clusters with `ccloud kafka cluster list`.
```
```
Error: no API key selected for resource "lkc-dvnr7"

Suggestions:
    Select an API key for resource "lkc-dvnr7" with `ccloud api-key use <API_KEY> --resource lkc-dvnr7`.
    To do so, you must have either already created or stored an API key for the resource.
    To create an API key use `ccloud api-key create --resource lkc-dvnr7`.
    To store an existing API key use `ccloud api-key store --resource lkc-dvnr7`.
```
```
Error: Kafka cluster "lkc-yydnp" not ready

Suggestions:
    It may take up to 5 minutes for a recently created Kafka cluster to be ready.
```

## Initializing the errors
There are four ways to create errors for the CLI.

1. All basic errors will fall under this category. Use one of the error intializing functions in the errors package (errors.New, errors.Errorf, errors.Wrap, errors.Wrapf), with an error message defined in [error_message.go](internal/pkg/errors/error_message.go). The name of the variable must end with *ErrorMsg*.
```
errors.Errorf(errors.AuthorizeAccountsErrorMsg, accountsStr)
```
2. For errors with suggestions, define the error message, and suggestions message next to each other in [error_message.go](internal/pkg/errors/error_message.go). The messages must have the same name with different ending following the naming convention (i.e. *ErrorMsg* and *Suggestions*). Then either `errors.NewErrorWithSuggestion` or `errors.NewWrapErrorWithSuggestion` is used to initialize the error.

```
ResolvingConfigPathErrorMsg        = "error resolving the config filepath at \"%s\" has occurred"
ResolvingConfigPathSuggestions     = "Try moving the config file to a different location."

return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.ResolvingConfigPathErrorMsg, c.Filename), errors.ResolvingConfigPathSuggestions)
```

```
LookUpRoleErrorMsg              = "failed to lookup role \"%s\""
LookUpRoleSuggestions           = "To check for valid roles, use `confluent role list`."

return errors.NewWrapErrorWithSuggestions(err, fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
```

3. If you know that your error will be used in many places, or need to be caught donwstream and process later, you can define typed error by implenting the `CLITypedError` interface
```
type CLITypedError interface {
	error
	UserFacingError() error
}
```
  See [typed.go](internal/pkg/errors/typed.go) for examples and interface definition.

4. For important errors that are thrown by external packages that need to be caught and translated, define a catcher in [catcher.go](internal/pkg/errors/catcher.go). The catcher can then be either inserted anyhwere in the CLI repo, or simply in `handleErrors` function in [handle.go](internal/pkg/errors/handle.go).

## HandleCommon
`errors.HandleCommon` is called for every cobra RunE or PrerunE command, to handle common logic required for all errors, including turning off the usage message. This is done under the hood when defining new CLI RunE and PrerunE with `cmd.NewCLIRunE` and `cmd.NewCLIPrerunE` intializers.

## Non-error communcation
For all non-error messages, define the string variables in [strings.go](internal/pkg/errors/strings.go), with variables name ending with *Msg*. The only exception being that warning messages are defined in [warning_message.go](internal/pkg/errors/warning_message.go) instead.
