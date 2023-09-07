package ai

import (
	"os"
	"strings"

	"github.com/charmbracelet/glamour"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"
	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type shell struct {
	client  *ccloudv2.Client
	history []aiv1.AiV1ChatCompletionsHistory
}

func (s *shell) executor(question string) {
	if s.isExit(question) {
		os.Exit(0)
	}

	req := aiv1.AiV1ChatCompletionsRequest{
		Question: aiv1.PtrString(question),
		History:  &s.history,
	}
	res, err := s.client.QueryChatCompletion(req)
	if err != nil {
		output.Printf("Error: %v", err)
		return
	}

	s.history = append(s.history, aiv1.AiV1ChatCompletionsHistory{
		Question: aiv1.PtrString(question),
		Answer:   aiv1.PtrString(res.GetAnswer()),
	})

	out, err := glamour.Render(res.GetAnswer(), "auto")
	if err != nil {
		output.Printf("Error: %v", err)
		return
	}

	output.Println(out)
}

func (s *shell) completer(d prompt.Document) []prompt.Suggest {
	return nil
}

func (s *shell) isExit(in string) bool {
	in = strings.ToLower(strings.TrimSpace(in))
	return in == "exit" || in == "quit"
}
