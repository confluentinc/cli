package ksql

import (
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

var subcommandFlags = map[string]*pflag.FlagSet{
	"create":         pcmd.ClusterEnvironmentContextSet(),
	"configure-acls": pcmd.ClusterEnvironmentContextSet(),
}
