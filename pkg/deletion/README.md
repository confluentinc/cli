## Supporting Multiple Resource Deletion

This document details how to implement `delete` commands which accept multiple arguments.

### Command Setup

Use `MinimumNArgs(1)` instead of `ExactArgs(1)`, and replace `validArgs` with a version supporting multiple autocompletion (named `validArgsMultiple`):

```go
func (c *command) newDeleteCommand() *cobra.Command {
    return &cobra.Command{
        Use:               "delete <id-1> [id-2] ... [id-n]",
        Args:              cobra.MinimumNArgs(1),
        ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
        RunE:              c.delete,
    }
}
```

Inside the `delete` function, first initialize required clients if applicable, and obtain any required information (flag values, the environment ID, the Kafka cluster ID, etc.):

```go
func (c *command) delete(cmd *cobra.Command, args []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}
}
```

### Validating Resource Existence

Next, we will validate the existence of all resources corresponding to the input argument IDs and prompt the user to confirm the delete operation.
The confirmation prompt is yes/no. Call it with `deletion.ValidateAndConfirm`.
The function signature is:

```go
func ValidateAndConfirm(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType) error
```

The function parameter takes in the resource ID string and returns a boolean value signifying whether or not the resource actually exists.
This usually requires a call to a `describe` SDK function. If the API doesn't support a READ operation, then you can use LIST instead.

```go
existenceFunc := func(id string) bool {
    _, _, err := c.V2Client.GetApiKey(id)
    return err == nil
}

if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.ApiKey); err != nil {
    return err
}
```

If you need to call the `describe` or `list` SDK functions for any other reason, store the results in a `map` or a `set` and use that inside the existenceFunc:

```go
connectorIdToName, err := c.mapConnectorIdToName(environmentId, kafkaCluster.ID)
if err != nil {
    return err
}

existenceFunc := func(id string) bool {
    _, ok := connectorIdToName[id]
    return ok
}

if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Connector); err != nil {
    return err
}
```

### Deleting the Resources

To perform the deletion, we call `deletion.Delete`. It's function signature is

```go
func Delete(args []string, callDeleteEndpoint func(string) error, resourceType string) ([]string, error)
```

This function will attempt to delete every resource provided in `args`. In other words, failure to delete one resource will not prevent `Delete` from attempting to delete the next resource.
The first return value is a list of IDs corresponding to successfully deleted resources. The error return value is a list of all errors encountered.

The function parameter takes the resource ID and calls the relevant `delete` SDK function.

For most resources, this is the last step and you can simply return the error to end the command:

```go
deleteFunc := func(id string) error {
    return c.V2Client.DeleteIamUser(id)
}

_, err = deletion.Delete(cmd, args, deleteFunc, resource.User)
return err
```

You can also wrap the errors returned by `deleteFunc`:

```go
deleteFunc := func(id string) error {
		if httpResp, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			return errors.CatchKafkaNotFoundError(err, id, httpResp)
		}
		return nil
}

_, err = deletion.Delete(cmd, args, deleteFunc, resource.User)
return err
```

For some resources, additional post-delete tasks need to be done. Use the first return value to process only the resources that were actually deleted.
Any errors resulting from these additional tasks should be appended to the error returned from `deletion.Delete` using `multierror`.
Note that you must call `ErrorOrNil()` for multierrors instead of returning them directly.

```go
deletedIds, err := deletion.Delete(cmd, args, deleteFunc, resource.ApiKey)

errs := multierror.Append(err, c.deleteKeysFromKeyStore(deletedIds))
```

Then, either return `errs.ErrorOrNil()` or process the error further:

```go
deletedIds, err := deletion.Delete(cmd, args, deleteFunc, resource.ApiKey)

errs := multierror.Append(err, c.deleteKeysFromKeyStore(deletedIds))
if errs.ErrorOrNil() != nil {
    return errors.NewErrorWithSuggestions(err.Error(), errors.APIKeyNotFoundSuggestions)
}

return nil
```

Lastly, if your resource is not immediately deleted, then you should instead call `deletion.DeleteWithoutMessage` and write your own custom deletion message instead:

```go
deletedIds, err := deletion.DeleteWithoutMessage(cmd, args, deleteFunc)
deleteMsg := "Started deletion of %s %s. To monitor a remove-broker task run `confluent kafka broker task list <id> --task-type remove-broker`.\n"
if len(deletedIds) == 1 {
    output.Printf(deleteMsg, resource.Broker, fmt.Sprintf(`"%s"`, deletedIds[0]))
} else if len(deletedIds) > 1 {
    output.Printf(deleteMsg, plural.Plural(resource.Broker), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
}

return err
```
