package auditlog

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var ConfigSubcommandFlags = map[string]*pflag.FlagSet{
	"config": cmd.ContextSet(),
}

var RouteSubcommandFlags = map[string]*pflag.FlagSet{
	"route": cmd.ContextSet(),
}
