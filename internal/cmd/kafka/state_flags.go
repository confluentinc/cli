package kafka

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var AclSubcommandFlags = map[string]*pflag.FlagSet{
	"acl": pcmd.ClusterEnvironmentContextSet(),
}

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"cluster": pcmd.EnvironmentContextSet(),
}

var GroupSubcommandFlags = map[string]*pflag.FlagSet{
	"consumer-group": pcmd.ClusterEnvironmentContextSet(),
}

var LagSubcommandFlags = map[string]*pflag.FlagSet{
	"lag": pcmd.ClusterEnvironmentContextSet(),
}

var TopicSubcommandFlags = map[string]*pflag.FlagSet{
	"topic": pcmd.ClusterEnvironmentContextSet(),
}

var LinkSubcommandFlags = map[string]*pflag.FlagSet{
	"link": pcmd.ClusterEnvironmentContextSet(),
}

var MirrorSubcommandFlags = map[string]*pflag.FlagSet{
	"mirror": pcmd.ClusterEnvironmentContextSet(),
}

var ProduceAndConsumeFlags = map[string]*pflag.FlagSet{
	"topic": pcmd.ClusterEnvironmentContextSet(),
}
