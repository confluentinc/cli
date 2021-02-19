package schema_registry

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"enable":   cmd.EnvironmentContextSet(),
	"describe": cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
	"update":   cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}

var SubjectSubcommandFlags = map[string]*pflag.FlagSet{
	"subject": cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}

var SchemaSubcommandFlags = map[string]*pflag.FlagSet{
	"schema": cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}

var OnPremClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"cluster": cmd.ContextSet(),
}
