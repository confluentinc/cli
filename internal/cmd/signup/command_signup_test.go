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
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"Confluent",
			"password",
			"y",
			"y",
			"y",
		),
		"A verification email has been sent to bstrauch@confluent.io.",
		"Success! Welcome to Confluent Cloud.",
	)
}

func TestSignupRejectTOS(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"Confluent",
			"password",
			"n", // Reject TOS
			"y", // Accept TOS after re-prompt
			"y",
			"y",
		),
		"You must accept to continue. To abandon flow, use Ctrl-C",
		"Success! Welcome to Confluent Cloud.",
	)
}

func TestSignupRejectPrivacyPolicy(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"Confluent",
			"password",
			"y",
			"n", // Reject PP
			"y", //Accept PP after re-prompt
			"y",
		),
		"You must accept to continue. To abandon flow, use Ctrl-C",
		"Success! Welcome to Confluent Cloud.",
	)
}

func TestSignupResendVerificationEmail(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"Confluent",
			"password",
			"y",
			"y",
			"n", // Resend
			"y", // Verify
		),
		"A verification email has been sent to bstrauch@confluent.io.",
		"A new verification email has been sent to bstrauch@confluent.io. If this email is not received, please contact support@confluent.io.",
		"Success! Welcome to Confluent Cloud.",
	)
}

func testSignup(t *testing.T, prompt pcmd.Prompt, expected ...string) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	err := signup(cmd, prompt, mockCcloudClient())
	require.NoError(t, err)

	for _, x := range expected {
		require.Contains(t, buf.String(), x)
	}
}

func mockCcloudClient() *ccloud.Client {
	return &ccloud.Client{
		Signup: &ccloudmock.Signup{
			CreateFunc: func(_ context.Context, _ *orgv1.SignupRequest) (*orgv1.SignupReply, error) {
				return nil, nil
			},
			SendVerificationEmailFunc: func(_ context.Context, _ *orgv1.Credentials) error {
				return nil
			},
		},
		Auth: &ccloudmock.Auth{
			LoginFunc: func(_ context.Context, _ string, _ string, _ string) (string, error) {
				return "", nil
			},
		},
	}
}
