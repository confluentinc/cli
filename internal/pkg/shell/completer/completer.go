package completer

import (
	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

type CompletionFunc = prompt.Completer
type CompleterFunc func(doc prompt.Document) []prompt.Suggest

type Completer interface {
	Complete(doc prompt.Document) []prompt.Suggest
}

type ServerCompletableCommand interface {
	Cmd() *cobra.Command
	ServerComplete() []prompt.Suggest
	ServerCompletableChildren() []*cobra.Command
}

type ServerCompletableFlag interface {
	Cmd() *cobra.Command
	ServerFlagComplete() map[string]func() []prompt.Suggest
	ServerCompletableFlagChildren() map[string][]*cobra.Command
}

type ServerSideCompleter interface {
	Completer
	AddCommand(cmd interface{})
	AddKafkaSubCommand(cmd interface{})
	AddStaticFlagCompletion(flagName string, suggestions []prompt.Suggest, commandPaths []string) // expose for testing
}

func (f CompleterFunc) Complete(doc prompt.Document) []prompt.Suggest {
	return f(doc)
}
