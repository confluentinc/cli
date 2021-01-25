package environment

import (
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

var SubcommandFlags = map[string]*pflag.FlagSet{
	"use":  cmd.ContextSet(),
	"list": cmd.ContextSet(),
}
