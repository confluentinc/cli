package schemaregistry

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"enable":   pcmd.EnvironmentContextSet(),
	"describe": pcmd.CombineFlagSet(pcmd.KeySecretSet(), pcmd.EnvironmentContextSet()),
	"update":   pcmd.CombineFlagSet(pcmd.KeySecretSet(), pcmd.EnvironmentContextSet()),
}

var SubjectSubcommandFlags = map[string]*pflag.FlagSet{
	"subject": pcmd.CombineFlagSet(pcmd.KeySecretSet(), pcmd.EnvironmentContextSet()),
}

var SchemaSubcommandFlags = map[string]*pflag.FlagSet{
	"schema": pcmd.CombineFlagSet(pcmd.KeySecretSet(), pcmd.EnvironmentContextSet()),
}

var ExporterSubcommandFlags = map[string]*pflag.FlagSet{
	"exporter": pcmd.CombineFlagSet(pcmd.KeySecretSet(), pcmd.EnvironmentContextSet()),
}
