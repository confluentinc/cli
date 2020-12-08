package cluster

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var ListFlags = map[string]*pflag.FlagSet{
	"list": cmd.ContextSet(),
}

var RegistryFlags = map[string]*pflag.FlagSet{
	"register": cmd.ContextSet(),
	"unregister": cmd.ContextSet(),
}
