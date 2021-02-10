package cluster

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"list":       cmd.ContextSet(),
	"register":   cmd.ContextSet(),
	"unregister": cmd.ContextSet(),
}
