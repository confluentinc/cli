package signup

import (
	"bytes"
	"context"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"

	cmdPkg "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	testToken  = "y0ur.jwt.T0kEn"
	promptUser = "prompt-user@confluent.io"
)

func TestSignupSuccess(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"US",
			"y",
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

func TestSignupBadCountryCode(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"ZZ",
			"US",
			"y",
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

func TestSignupRejectCountryCode(t *testing.T) {
	testSignup(t,
		mock.NewPromptMock(
			"bstrauch@confluent.io",
			"Brian",
			"Strauch",
			"CH",
			"n",
			"US",
			"y",
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
			"US",
			"y",
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
			"US",
			"y",
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
			"US",
			"y",
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

func testSignup(t *testing.T, prompt form.Prompt, expected ...string) {
	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	signupCmd := newCmd(v3.AuthenticatedCloudConfigMock())
	signupCmd.Config = &cmdPkg.DynamicConfig{
		Config: v3.UnauthenticatedCloudConfigMock(),
	}

	err := signupCmd.Signup(cmd, prompt, mockCcloudClient())
	require.NoError(t, err)

	for _, x := range expected {
		require.Contains(t, buf.String(), x)
	}
}

func newCmd(conf *v3.Config) *command {
	client := mockCcloudClient()
	prerunner := cliMock.NewPreRunnerMock(client, nil, nil, conf)
	auth := &ccloudmock.Auth{
		LoginFunc: func(ctx context.Context, idToken string, username string, password string) (string, error) {
			return testToken, nil
		},
		UserFunc: func(ctx context.Context) (*orgv1.GetUserReply, error) {
			return &orgv1.GetUserReply{
				User: &orgv1.User{
					Id:        23,
					Email:     promptUser,
					FirstName: "Cody",
				},
				Accounts: []*orgv1.Account{{Id: "a-595", Name: "Default"}},
			}, nil
		},
	}
	user := &ccloudmock.User{}
	ccloudClientFactory := &cliMock.MockCCloudClientFactory{
		JwtHTTPClientFactoryFunc: func(ctx context.Context, jwt, baseURL string) *ccloud.Client {
			return &ccloud.Client{Auth: auth, User: user}
		},
	}
	cmd := New(prerunner, conf.Logger, "ccloud-cli", ccloudClientFactory)
	return cmd
}

func mockCcloudClient() *ccloud.Client {
	return &ccloud.Client{
		Signup: &ccloudmock.Signup{
			CreateFunc: func(_ context.Context, _ *orgv1.SignupRequest) (*orgv1.SignupReply, error) {
				return nil, nil
			},
			SendVerificationEmailFunc: func(_ context.Context, _ *orgv1.User) error {
				return nil
			},
		},
		Auth: &ccloudmock.Auth{
			LoginFunc: func(_ context.Context, _ string, _ string, _ string) (string, error) {
				return "", nil
			},
		},
		User:   &ccloudmock.User{},
		Params: &ccloud.Params{BaseURL: "http://baseurl.com"},
	}
}
