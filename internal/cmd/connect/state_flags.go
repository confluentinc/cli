package connect

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var subcommandFlags = map[string]*pflag.FlagSet{
	"describe": pcmd.ClusterEnvironmentContextSet(),
	"list":     pcmd.ClusterEnvironmentContextSet(),
	"create":   pcmd.ClusterEnvironmentContextSet(),
	"delete":   pcmd.ClusterEnvironmentContextSet(),
	"update":   pcmd.ClusterEnvironmentContextSet(),
	"pause":    pcmd.ClusterEnvironmentContextSet(),
	"resume":   pcmd.ClusterEnvironmentContextSet(),
}
