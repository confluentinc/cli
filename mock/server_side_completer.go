package mock

import (
	"github.com/c-bata/go-prompt"
)

type ServerSideCompleter struct {
}

func (*ServerSideCompleter) Complete(doc prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func (*ServerSideCompleter) AddCommand(cmd interface{}) {

}

func (*ServerSideCompleter) AddStaticFlagCompletion(flagName string, suggestions []prompt.Suggest, commandPaths []string) {

}
