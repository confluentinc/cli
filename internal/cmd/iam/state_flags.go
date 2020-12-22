package iam

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var AclSubcommandFlags = map[string]*pflag.FlagSet{
	"acl": cmd.ContextSet(),
}

var RoleSubcommandFlags = map[string]*pflag.FlagSet{
	"role": cmd.ContextSet(),
}

var RolebindingSubcommandFlags = map[string]*pflag.FlagSet{
	"rolebinding": cmd.ContextSet(),
}
