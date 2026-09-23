package pair

import (
	"testing"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"
	"github.com/stretchr/testify/require"
)

func condition(member, typ, status, reason, message string) switchoverv1.SwitchoverV1Condition {
	c := switchoverv1.SwitchoverV1Condition{Type: typ, Status: status}
	if member != "" {
		c.Member = switchoverv1.PtrString(member)
	}
	if reason != "" {
		c.Reason = switchoverv1.PtrString(reason)
	}
	if message != "" {
		c.Message = switchoverv1.PtrString(message)
	}
	return c
}

func TestFormatConditions(t *testing.T) {
	t.Run("member leads the line so a two-member pair is unambiguous", func(t *testing.T) {
		got := formatConditions([]switchoverv1.SwitchoverV1Condition{
			condition("dr-az-pub-eus", "ResourceUnplannedFailoverComplete", "True", "AllTopicsTransitioned", "10/10 topics completed"),
			condition("dr-az-pub-cus", "ResourceUnplannedFailoverComplete", "False", "TopicsTransitioning", "3/10 topics completed"),
		})
		require.Equal(t,
			"Member=dr-az-pub-eus ResourceUnplannedFailoverComplete=True (AllTopicsTransitioned): 10/10 topics completed\n"+
				"Member=dr-az-pub-cus ResourceUnplannedFailoverComplete=False (TopicsTransitioning): 3/10 topics completed",
			got)
	})

	t.Run("no member keeps the original shape", func(t *testing.T) {
		got := formatConditions([]switchoverv1.SwitchoverV1Condition{
			condition("", "ResourceValidationComplete", "False", "ValidationFailed", "topic is not writeable"),
		})
		require.Equal(t, "ResourceValidationComplete=False (ValidationFailed): topic is not writeable", got)
	})

	t.Run("reason and message are optional", func(t *testing.T) {
		got := formatConditions([]switchoverv1.SwitchoverV1Condition{condition("west", "Ready", "True", "", "")})
		require.Equal(t, "Member=west Ready=True", got)
	})

	t.Run("empty", func(t *testing.T) {
		require.Equal(t, "", formatConditions(nil))
	})
}
