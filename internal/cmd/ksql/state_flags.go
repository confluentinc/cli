package ksql

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var subcommandFlags = map[string]*pflag.FlagSet{
	"list":           pcmd.EnvironmentContextSet(),
	"create":         pcmd.ClusterEnvironmentContextSet(),
	"describe":       pcmd.EnvironmentContextSet(),
	"delete":         pcmd.EnvironmentContextSet(),
	"configure-acls": pcmd.ClusterEnvironmentContextSet(),
}
