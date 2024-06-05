package ai

import (
	"time"

	"github.com/google/uuid"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2/ai/v1"
)

type session struct {
	id        string
	history   []aiv1.AiV1ChatCompletionsHistory
	expiresAt time.Time
}

func newSession() *session {
	return &session{
		id:        uuid.New().String(),
		history:   []aiv1.AiV1ChatCompletionsHistory{},
		expiresAt: time.Now().Add(time.Hour),
	}
}

func (s *session) addHistory(history aiv1.AiV1ChatCompletionsHistory) {
	s.history = append(s.history, history)
	if len(s.history) > 5 {
		s.history = s.history[:5]
	}
}

func (s *session) isExpired() bool {
	return time.Now().After(s.expiresAt)
}
