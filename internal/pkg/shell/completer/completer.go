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

type ServerSideCompleter interface {
	Completer
	AddCommand(cmd ServerCompletableCommand)
}

func (f CompleterFunc) Complete(doc prompt.Document) []prompt.Suggest {
	return f(doc)
}
