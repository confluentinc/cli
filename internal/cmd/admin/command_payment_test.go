package admin

import (
	"bytes"
	"context"
	"testing"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/mock"
	climock "github.com/confluentinc/cli/mock"
)

func TestPaymentDescribe(t *testing.T) {
	cmd := mockAdminCommand()

	out, err := pcmd.ExecuteCommand(cmd, "payment", "describe")
	require.NoError(t, err)
	require.Equal(t, "Visa ending in 4242\n", out)
}

type PaymentUpdateSuite struct {
	prompt   *mock.Prompt
	expected []string
}

func TestPaymentUpdate(t *testing.T) {
	c := getCommand()
	cmd := mockAdminCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	tests := []*PaymentUpdateSuite{
		&PaymentUpdateSuite{
			prompt: mock.NewPromptMock(
				"4242424242424242",
				"12/70",
				"999",
				"Brian Strauch",
			),
			expected: []string{"Updated"},
		},
	}

	for _, test := range tests {
		err := c.update(cmd, test.prompt)
		for _, expectedOutput := range test.expected {
			require.Contains(t, buf.String(), expectedOutput)
		}
		require.NoError(t, err)
	}
}

func TestPaymentRegexValidation(t *testing.T) {
	c := getCommand()
	cmd := mockAdminCommand()
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	tests := []*PaymentUpdateSuite{
		&PaymentUpdateSuite{
			prompt: mock.NewPromptMock(
				"42424242",                 //too short
				"424242424242424242424242", //too long
				"4242424242a42",            //non-digit characters
				"4242424242424242",
				"12/70",
				"999",
				"Brian Strauch",
			),
			expected: []string{
				"\"42424242\" is not of valid format for field \"card number\"",
				"\"424242424242424242424242\" is not of valid format for field \"card number\"",
				"\"4242424242a42\" is not of valid format for field \"card number\"",
				"Updated.",
			},
		},
		&PaymentUpdateSuite{
			prompt: mock.NewPromptMock(
				"4242424242424242",
				"121/70", //too many digits for month
				"12/701", //too many digits for year
				"aa/70",  //non-digit characters
				"1270",   //no /
				"12/70",
				"999",
				"Brian Strauch",
			),
			expected: []string{
				"\"121/70\" is not of valid format for field \"expiration\"",
				"\"12/701\" is not of valid format for field \"expiration\"",
				"\"aa/70\" is not of valid format for field \"expiration\"",
				"\"1270\" is not of valid format for field \"expiration\"",
				"Updated.",
			},
		},
		&PaymentUpdateSuite{
			prompt: mock.NewPromptMock(
				"4242424242424242",
				"12/70",
				"999999", //too long
				"99",     //too short
				"999a",   //non-digit characters
				"999",
				"Brian Strauch",
			),
			expected: []string{
				"\"999999\" is not of valid format for field \"cvc\"",
				"\"99\" is not of valid format for field \"cvc\"",
				"\"999a\" is not of valid format for field \"cvc\"",
				"Updated.",
			},
		},
	}
	for _, test := range tests {
		err := c.update(cmd, test.prompt)
		for _, expectedOutput := range test.expected {
			require.Contains(t, buf.String(), expectedOutput)
		}
		require.NoError(t, err)
	}
}

func getCommand() (c *command) {
	c = &command{
		AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{
			CLICommand: &pcmd.CLICommand{
				Command: mockAdminCommand(),
				Config:  nil,
				Version: nil,
			},
			Client: mockClient(),
			State: &v2.ContextState{
				Auth: &v1.AuthConfig{
					User: &orgv1.User{
						OrganizationId: int32(0),
					},
				},
			},
		},
		isTest: true,
	}
	return
}

func mockAdminCommand() *cobra.Command {
	client := mockClient()
	cfg := v3.AuthenticatedCloudConfigMock()
	return New(climock.NewPreRunnerMock(client, nil, cfg), true)
}

func mockClient() (client *ccloud.Client) {
	client = &ccloud.Client{
		Organization: &ccloudmock.Organization{
			GetPaymentInfoFunc: func(_ context.Context, _ *orgv1.Organization) (*orgv1.Card, error) {
				card := &orgv1.Card{
					Brand: "Visa",
					Last4: "4242",
				}
				return card, nil
			},
			UpdatePaymentInfoFunc: func(_ context.Context, _ *orgv1.Organization, _ string) error {
				return nil
			},
		},
	}
	return
}
