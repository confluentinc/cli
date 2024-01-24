package ai

import (
	"time"

	aiv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/ai/v1"
	"github.com/google/uuid"
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

func (s *session) isExpired() bool {
	return time.Now().After(s.expiresAt)
}
