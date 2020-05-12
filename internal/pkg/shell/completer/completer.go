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

type CommandCompleter interface {
	Completer
	AddCommand(cmd *cobra.Command, completionFunc CompletionFunc)
}

func (f CompleterFunc) Complete(doc prompt.Document) []prompt.Suggest {
	return f(doc)
}
