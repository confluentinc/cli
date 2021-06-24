package connectorcatalog

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"connector-catalog": cmd.ClusterEnvironmentContextSet(),
}
