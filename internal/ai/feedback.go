package ai

import (
	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"
	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
)

// Feedback is collected across three interactions:
// 1. A question is answered and the question ID and session ID are stored
// 2. The user reacts to the answer with +1 or -1
// 3. The user provides a comment or skips with enter
type feedback struct {
	sessionId string
	answerId  string
	reaction  string
	comment   string
}

func newFeedback(answerId string) *feedback {
	return &feedback{answerId: answerId}
}

func (f *feedback) setReaction(reaction string) {
	if reaction == "+1" {
		f.reaction = "THUMBS_UP"
	} else {
		f.reaction = "THUMBS_DOWN"
	}
}

func (f *feedback) create(client *ccloudv2.Client, sessionId string) error {
	feedback := aiv1.AiV1Feedback{
		AiSessionId: &sessionId,
		Reaction:    &f.reaction,
		Comment:     *aiv1.NewNullableString(&f.comment),
	}

	return client.CreateChatCompletionFeedback(f.answerId, feedback)
}
