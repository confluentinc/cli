package auditlog

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var ConfigSubcommandFlags = map[string]*pflag.FlagSet{
	"config": cmd.ContextSet(),
}

var RouteSubcommandFlags = map[string]*pflag.FlagSet{
	"route": cmd.ContextSet(),
}
