package mock

import (
	"github.com/c-bata/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/shell/completer"
)

type ServerSideCompleter struct {
}

func (*ServerSideCompleter) Complete(doc prompt.Document) []prompt.Suggest {
	return []prompt.Suggest{}
}
func (*ServerSideCompleter) AddCommand(cmd completer.ServerCompletableCommand) {

}
