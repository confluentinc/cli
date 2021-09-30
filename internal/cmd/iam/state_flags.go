package iam

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var AclSubcommandFlags = map[string]*pflag.FlagSet{
	"acl": cmd.ContextSet(),
}

var RoleSubcommandFlags = map[string]*pflag.FlagSet{
	"role": cmd.ContextSet(),
}

var RoleBindingSubcommandFlags = map[string]*pflag.FlagSet{
	"role-binding": cmd.ContextSet(),
}
