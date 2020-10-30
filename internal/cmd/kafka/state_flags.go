package kafka

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet {
	"cluster"	:	cmd.EnvironmentContextSet(),
}

var AclSubcommandFlags = map[string]*pflag.FlagSet {
	"acl"	:	cmd.ClusterEnvironmentContextSet(),
}

var TopicSubcommandFlags = map[string]*pflag.FlagSet {
	"topic"	:	cmd.ClusterEnvironmentContextSet(),
}

var LinkSubcommandFlags = map[string]*pflag.FlagSet {
	"link"	:	cmd.ClusterEnvironmentContextSet(),
}
