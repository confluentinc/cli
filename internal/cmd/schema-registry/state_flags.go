package schemaregistry

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"enable":   pcmd.EnvironmentContextSet(),
	"describe": pcmd.EnvironmentContextSet(),
	"update":   pcmd.EnvironmentContextSet(),
}

var SubjectSubcommandFlags = map[string]*pflag.FlagSet{
	"subject": pcmd.EnvironmentContextSet(),
}

var SchemaSubcommandFlags = map[string]*pflag.FlagSet{
	"schema": pcmd.EnvironmentContextSet(),
}

var ExporterSubcommandFlags = map[string]*pflag.FlagSet{
	"exporter": pcmd.EnvironmentContextSet(),
}
