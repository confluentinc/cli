package cluster

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"list": cmd.ContextSet(),
	"register": cmd.ContextSet(),
	"unregister": cmd.ContextSet(),
}
