package schema_registry

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet {
	"enable"	:	cmd.EnvironmentContextSet(),
	"describe"	:	cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
	"update"	:	cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}

var SubjectSubcommandFlags = map[string]*pflag.FlagSet {
	"acl" : cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}

var SchemaSubcommandFlags = map[string]*pflag.FlagSet {
	"schema"	:	cmd.CombineFlagSet(cmd.KeySecretSet(), cmd.EnvironmentContextSet()),
}
