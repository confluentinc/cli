package completer

import (
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
)

func TestConditionalWrapper(t *testing.T) {
	c := CompleterFunc(func(d prompt.Document) []prompt.Suggest {
		fmt.Println("Hello!")
		return nil
	})
	logWrapper := LogWrapper("logging!")
	cond := true
	condCompleter := ConditionalWrapper(&cond, logWrapper)(c)
	condCompleter.Complete(*prompt.NewDocument())
	cond = false
	condCompleter.Complete(*prompt.NewDocument())
	cond = true
	condCompleter.Complete(*prompt.NewDocument())
}
