package logs

// Error messages and suggestions for `confluent logs`. Messages are short, lowercase and unpunctuated; anything
// the user can do about them goes in the matching suggestions, per pkg/errors/README.md.
const (
	sinceWithFromToErrorMsg  = "`--since` cannot be used with `--from` or `--to`"
	missingWindowErrorMsg    = "one of `--since` or `--from` is required"
	missingWindowSuggestions = "Use `--since` for a window ending now, for example `--since 1h`, or `--from` for an absolute start time."
	toBeforeFromErrorMsg     = "`--to` must not be before `--from`"

	invalidSinceErrorMsg    = "invalid `--since` value \"%s\""
	invalidSinceSuggestions = "Use a positive duration such as 1h, 24h or 7d."

	invalidTimestampErrorMsg    = "invalid `%s` value \"%s\""
	invalidTimestampSuggestions = "Use an ISO 8601 UTC timestamp such as 2026-08-20T00:00:00Z."

	tailOutOfRangeErrorMsg = "`--tail` must be between 1 and %d"

	invalidSearchErrorMsg         = "invalid `--search` value \"%s\""
	invalidSearchSuggestions      = "Use the form \"field=value\", for example `--search level=ERROR`."
	reservedSearchFieldErrorMsg   = "`--search` cannot set the reserved field \"search_text\""
	reservedSearchFieldSuggestion = "Use `--search-text` for full-text search."

	searchTextUnsupportedErrorMsg = "\"%s\" logs do not support `--search-text`"

	unknownSearchFieldErrorMsg    = "\"%s\" is not a search field for \"%s\" logs"
	unknownSearchFieldSuggestions = "Available search fields: %s.\nRun `confluent logs source describe %s` for details."

	invalidSearchValueErrorMsg    = "\"%s\" is not a valid value for \"%s\""
	invalidSearchValueSuggestions = "Allowed values: %s."

	resourceIdRequiredErrorMsg    = "log source \"%s\" requires a resource ID"
	resourceIdRequiredSuggestions = "Pass the resource ID as an argument, for example `confluent logs fetch <resource-id> --source %s`."

	unsupportedScopeErrorMsg    = "log source \"%s\" does not support %s scope"
	unsupportedScopeSuggestions = "Supported scopes: %s."

	resourceIdMismatchErrorMsg    = "resource ID \"%s\" does not match log source \"%s\""
	resourceIdMismatchSuggestions = "This log source expects IDs starting with \"%s-\"."

	lookbackExceededErrorMsg    = "`--since` exceeds the %dd lookback for \"%s\""
	lookbackExceededSuggestions = "Use a window of at most %dd."

	unsupportedCrnElementErrorMsg    = "log source \"%s\" uses a CRN element \"%s\" that this version of the CLI does not support"
	unsupportedCrnElementSuggestions = "Upgrade the CLI to the latest version."

	missingScopeElementErrorMsg = "log source \"%s\" does not declare an %s element to scope a query to"

	environmentNotFoundErrorMsg    = "environment ID not found"
	environmentNotFoundSuggestions = "Set an environment with `confluent environment use <env-id>` or specify `--environment`."

	kafkaClusterNotFoundErrorMsg    = "Kafka cluster information not found"
	kafkaClusterNotFoundSuggestions = "Select a cluster with `confluent kafka cluster use <cluster-id>` or specify `--cluster`."
)
