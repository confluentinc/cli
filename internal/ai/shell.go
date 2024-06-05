package ai

import (
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"
	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type shell struct {
	client           *ccloudv2.Client
	session          *session
	feedback         *feedback
	cleanupFunctions []func()
}

func newShell(client *ccloudv2.Client) *shell {
	return &shell{
		client:  client,
		session: newSession(),
	}
}

func (s *shell) AddCleanupFunction(cleanupFunction func()) *shell {
	s.cleanupFunctions = append(s.cleanupFunctions, cleanupFunction)
	return s
}

func (s *shell) executor(input string) {
	input = strings.TrimSpace(input)

	if input == "exit" {
		s.cleanup()
		os.Exit(0)
	}

	if s.feedback != nil && (s.feedback.reaction != "" || input == "+1" || input == "-1") {
		if s.feedback.reaction != "" {
			s.feedback.comment = input

			if err := s.feedback.create(s.client, s.session.id); err != nil {
				s.exitWithErr(err)
			}

			s.feedback = nil
			output.Println(false, "Thanks for your feedback!")

			return
		}

		s.feedback.setReaction(input)
		output.Println(false, "Provide more details below or press enter to skip.")

		return
	}

	m := &model{}

	if s.session.isExpired() {
		s.session = newSession()
	}

	go func() {
		req := aiv1.AiV1ChatCompletionsRequest{
			AiSessionId:  &s.session.id,
			Question:     &input,
			DriftEnabled: aiv1.PtrBool(true),
			History:      &s.session.history,
		}

		reply, err := s.client.QueryChatCompletion(req)
		if err != nil {
			s.exitWithErr(err)
		}

		s.feedback = newFeedback(reply.GetId())
		s.session.expiresAt = time.Now().Add(time.Hour)
		s.session.addHistory(aiv1.AiV1ChatCompletionsHistory{
			Question: aiv1.PtrString(input),
			Answer:   aiv1.PtrString(reply.GetAnswer()),
		})

		out, err := glamour.Render(reply.GetAnswer(), "auto")
		if err != nil {
			s.exitWithErr(err)
		}
		m.out = out
	}()

	if _, err := tea.NewProgram(m).Run(); err != nil {
		s.exitWithErr(err)
	}

	// Print outside of bubbletea to avoid text cropping
	if m.out != "" {
		output.Print(false, m.out)
	}
}

func (s *shell) cleanup() {
	for _, cleanupFunction := range s.cleanupFunctions {
		cleanupFunction()
	}
}

func (s *shell) completer(_ prompt.Document) []prompt.Suggest {
	return nil
}

func (s *shell) exitWithErr(err error) {
	s.cleanup()
	output.Printf(false, "Error: %v\n", err)
	os.Exit(1)
}
