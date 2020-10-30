package environment

import (
	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/spf13/pflag"
)

var SubcommandFlags = map[string]*pflag.FlagSet {
	"use"	:	cmd.ContextSet(),
	"list"	:	cmd.ContextSet(),
}
