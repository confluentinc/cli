package schemaregistry

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"describe": pcmd.KeySecretSet(),
	"update":   pcmd.KeySecretSet(),
}

var SubjectSubcommandFlags = map[string]*pflag.FlagSet{
	"subject": pcmd.KeySecretSet(),
}

var SchemaSubcommandFlags = map[string]*pflag.FlagSet{
	"schema": pcmd.KeySecretSet(),
}

var ExporterSubcommandFlags = map[string]*pflag.FlagSet{
	"exporter": pcmd.KeySecretSet(),
}
