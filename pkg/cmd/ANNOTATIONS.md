## Cloud and On-Prem Annotations

Occasionally, we want to restrict commands to be used only when the user is logged in to Confluent Cloud or Confluent Platform.
We use [Cobra](https://github.com/spf13/cobra) annotations to label commands as Cloud-only or On-Prem-only.
Applying an annotation to a parent command will recursively apply the same annotation to its children.

For example, the `confluent billing` commands have been labeled as Cloud-only:

```go
cmd := &cobra.Command{
    Use:         "billing",
    Short:       "Perform billing administrative tasks for the current organization.",
    Args:        cobra.NoArgs,
    Annotations: map[string]string{annotations.RunRequirement: annotations.RequireCloudLogin},
}
```

Trying to use any `confluent billing` command results in the following error:

    $ confluent billing payment describe
    Error: you must log in to Confluent Cloud to use this command
