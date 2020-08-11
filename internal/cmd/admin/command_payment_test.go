package admin

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
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/mock"
)

func TestPaymentDescribe(t *testing.T) {
	cmd := mockAdminCommand()

	out, err := pcmd.ExecuteCommand(cmd, "payment", "describeRunE")
	require.NoError(t, err)
	require.Equal(t, "Visa ending in 4242\n", out)
}

func TestPaymentUpdate(t *testing.T) {
	c := command{
		AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{
			State: &v2.ContextState{
				Auth: &v1.AuthConfig{
					User: &orgv1.User{
						OrganizationId: int32(0),
					},
				},
			},
		},
	}

	cmd := &cobra.Command{}
	buf := new(bytes.Buffer)
	cmd.SetOut(buf)

	prompt := mock.NewPromptMock(
		"4242424242424242",
		"12/70",
		"999",
		"Brian Strauch",
	)

	err := c.update(cmd, prompt)
	require.NoError(t, err)
	require.Contains(t, "Updated.", buf.String())
}

func mockAdminCommand() *cobra.Command {
	client := &ccloud.Client{
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

	cfg := v3.AuthenticatedCloudConfigMock()

	return New(mock.NewPreRunnerMock(client, nil, cfg))
}
