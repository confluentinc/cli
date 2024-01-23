# Output

## User-facing strings

User-facing strings appear in command and flag descriptions, examples, and error messages and should be formatted consistently.

1. If the string contains a CLI command or flag, it must be formatted with backticks.
```
suggestion := "Run `confluent kafka cluster list` to list the Kafka clusters in the current environment."
```

2. If the string contains a resource (anything that a user might type that is not a CLI command or flag), it must be formatted with quotes.
```
suggestion := `Update Kafka cluster "lkc-123456".`
```

3. If the string contains both (1) and (2), format it in the following way.
```
suggestion := "Update Kafka cluster \"lkc-123456\" with `confluent kafka cluster update lkc-123456`."
```

Note: Due to a limitation of the spf13/pflag package, flag descriptions must never contain backticks, so quotes should be used for flags, commands, and resources.
