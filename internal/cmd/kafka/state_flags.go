package kafka

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var AclSubcommandFlags = map[string]*pflag.FlagSet{
	"acl": cmd.ClusterEnvironmentContextSet(),
}

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"cluster": cmd.EnvironmentContextSet(),
}

var GroupSubcommandFlags = map[string]*pflag.FlagSet{
	"consumer-group": cmd.ClusterEnvironmentContextSet(),
}

var LagSubcommandFlags = map[string]*pflag.FlagSet{
	"lag": cmd.ClusterEnvironmentContextSet(),
}

var TopicSubcommandFlags = map[string]*pflag.FlagSet{
	"topic": cmd.ClusterEnvironmentContextSet(),
}

var LinkSubcommandFlags = map[string]*pflag.FlagSet{
	"link": cmd.ClusterEnvironmentContextSet(),
}

var ProduceAndConsumeFlags = map[string]*pflag.FlagSet{
	"topic": cmd.CombineFlagSet(cmd.ClusterEnvironmentContextSet(), cmd.KeySecretSet()),
}

var OnPremClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"cluster": cmd.ContextSet(),
}
