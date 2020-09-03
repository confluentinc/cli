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

type CompletableCommand interface {
	Cmd() *cobra.Command
	Complete() []prompt.Suggest
	CompletableChildren() []*cobra.Command
}

func (f CompleterFunc) Complete(doc prompt.Document) []prompt.Suggest {
	return f(doc)
}
