package apikey

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var SubcommandFlags = map[string]*pflag.FlagSet {
	"create" :	cmd.EnvironmentContextSet(),
	"store" : cmd.EnvironmentContextSet(),
}
