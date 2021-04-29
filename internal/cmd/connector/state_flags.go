package connector

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"describe": cmd.ClusterEnvironmentContextSet(),
	"list":     cmd.ClusterEnvironmentContextSet(),
	"create":   cmd.ClusterEnvironmentContextSet(),
	"delete":   cmd.ClusterEnvironmentContextSet(),
	"update":   cmd.ClusterEnvironmentContextSet(),
	"pause":    cmd.ClusterEnvironmentContextSet(),
	"resume":   cmd.ClusterEnvironmentContextSet(),
}
