package connect

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var ClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"list": cmd.ContextSet(),
}
