package ai

import (
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"
	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type shell struct {
	client  *ccloudv2.Client
	session *session
}

func (s *shell) executor(question string) {
	if s.isExit(question) {
		os.Exit(0)
	}

	m := &model{}

	if s.session.isExpired() {
		s.session = newSession()
	}

	go func() {
		req := aiv1.AiV1ChatCompletionsRequest{
			AiSessionId: &s.session.id,
			Question:    &question,
			History:     &s.session.history,
		}

		res, err := s.client.QueryChatCompletion(req)
		if err != nil {
			exitWithErr(err)
		}

		s.session.history = append(s.session.history, aiv1.AiV1ChatCompletionsHistory{
			Question: aiv1.PtrString(question),
			Answer:   aiv1.PtrString(res.GetAnswer()),
		})

		out, err := glamour.Render(res.GetAnswer(), "auto")
		if err != nil {
			exitWithErr(err)
		}
		m.out = out
	}()

	if _, err := tea.NewProgram(m).Run(); err != nil {
		exitWithErr(err)
	}

	// Print outside of bubbletea to avoid text cropping
	if m.out != "" {
		output.Print(false, m.out)
	}
}

func (s *shell) completer(_ prompt.Document) []prompt.Suggest {
	return nil
}

func (s *shell) isExit(in string) bool {
	return strings.TrimSpace(in) == "exit"
}

func exitWithErr(err error) {
	output.Printf(false, "Error: %v\n", err)
	os.Exit(0)
}
