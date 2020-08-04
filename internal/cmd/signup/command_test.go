package signup

import (
	"bytes"
	"context"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/mock"
)

func TestSignupSuccess(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"Brian",
			"Strauch",
			"Confluent",
			"bstrauch@confluent.io",
			"pass",
			"y",
			"y",
			"y",
		),
		"Success! Welcome to Confluent Cloud.",
	)
}

func TestSignupRejectTOS(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"Brian",
			"Strauch",
			"Confluent",
			"bstrauch@confluent.io",
			"pass",
			"n",
			"y",
			"y",
		),
		"You must accept the Terms of Service.",
	)
}

func TestSignupRejectPrivacyPolicy(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"Brian",
			"Strauch",
			"Confluent",
			"bstrauch@confluent.io",
			"pass",
			"y",
			"n",
			"y",
		),
		"You must accept the Privacy Policy.",
	)
}

func testSignup(t *testing.T, prompt pcmd.Prompt, expected string) {
	c := &command{}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := c.signup(cmd, prompt, mockCcloudClient())
	require.NoError(t, err)

	require.Contains(t, buf.String(), expected)
}

func mockCcloudClient() *ccloud.Client {
	return &ccloud.Client{
		Signup: &ccloudmock.Signup{
			CreateFunc: func(_ context.Context, req *orgv1.SignupRequest) (*orgv1.SignupReply, error) {
				return nil, nil
			},
		},
	}
}
