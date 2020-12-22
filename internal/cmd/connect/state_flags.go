package connect

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"list": cmd.ContextSet(),
}
