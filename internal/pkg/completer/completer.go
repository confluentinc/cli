package completer

import (
	"fmt"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"
)

var (
	UnauthenticatedSuggestion = []prompt.Suggest{
		{
			Text:        " ",
			Description: "You are currently not authenticated. Please login.",
		},
	}
)

type CompletionFunc = prompt.Completer
type CommandCompletionFunc func() []prompt.Suggest
type CompleterFunc func(d prompt.Document) []prompt.Suggest
type CompleterWrapper func(Completer) Completer

type Completer interface {
	Complete(d prompt.Document) []prompt.Suggest
}

type CommandCompleter interface {
	Completer
	AddCommand(cmd *cobra.Command, completionFunc CommandCompletionFunc)
}

func (f CompleterFunc) Complete(d prompt.Document) []prompt.Suggest {
	return f(d)
}

func LogWrapper(s string) CompleterWrapper {
	return func(completer Completer) Completer {
		return CompleterFunc(func(d prompt.Document) []prompt.Suggest {
			fmt.Println(s)
			return completer.Complete(d)
		})
	}
}

func ConditionalWrapper(condition *bool, wrapper CompleterWrapper) CompleterWrapper {
	return func(completer Completer) Completer {
		return CompleterFunc(func(d prompt.Document) []prompt.Suggest {
			if *condition {
				return wrapper(completer).Complete(d)
			}
			return completer.Complete(d)
		})
	}
}
