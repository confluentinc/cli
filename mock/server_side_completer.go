package mock

import (
	"github.com/c-bata/go-prompt"
)

type ServerSideCompleter struct {
}

func (*ServerSideCompleter) Complete(_ prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}

func (*ServerSideCompleter) AddCommand(_ interface{}) {

}

func (*ServerSideCompleter) AddStaticFlagCompletion(_ string, _ []prompt.Suggest, _ []string) {

}
