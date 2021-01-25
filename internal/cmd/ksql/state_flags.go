package ksql

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"list":           cmd.EnvironmentContextSet(),
	"create":         cmd.ClusterEnvironmentContextSet(),
	"describe":       cmd.EnvironmentContextSet(),
	"delete":         cmd.EnvironmentContextSet(),
	"configure-acls": cmd.ClusterEnvironmentContextSet(),
}

var OnPremClusterSubcommandFlags = map[string]*pflag.FlagSet{
	"cluster":        cmd.ContextSet(),
}
