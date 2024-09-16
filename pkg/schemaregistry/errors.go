package schemaregistry

import "github.com/confluentinc/cli/v3/pkg/errors"

var ErrNotEnabled = errors.NewErrorWithSuggestions(
	"Schema Registry not enabled",
	"Enable Schema Registry for this environment by creating a Kafka cluster with `confluent kafka cluster create`.",
)
