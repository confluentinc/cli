package auditlog

import (
	"context"
	"testing"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestAuditLogDescribe(t *testing.T) {
	t.Parallel()

	cmd := mockAuditLogCommand(true)

	_, err := pcmd.ExecuteCommand(cmd, "describe")
	require.NoError(t, err)
}

func TestAuditLogDescribeUnconfigured(t *testing.T) {
	t.Parallel()

	cmd := mockAuditLogCommand(false)

	_, err := pcmd.ExecuteCommand(cmd, "describe")
	require.Error(t, err)
	require.Equal(t, errors.AuditLogsNotEnabledErrorMsg, err.Error())
}

func mockAuditLogCommand(configured bool) *cobra.Command {
	client := &ccloudv1.Client{
		User: &ccloudv1mock.UserInterface{
			GetServiceAccountFunc: func(_ context.Context, id int32) (*ccloudv1.User, error) {
				return &ccloudv1.User{ResourceId: "sa-123456"}, nil
			},
		},
	}

	cfg := v1.AuthenticatedCloudConfigMock()

	if configured {
		cfg.Context().State.Auth.Organization.AuditLog = &ccloudv1.AuditLog{
			ClusterId:        "lkc-ab123",
			AccountId:        "env-zy987",
			ServiceAccountId: 12345,
			TopicName:        "confluent-audit-log-events",
		}
	}

	return New(climock.NewPreRunnerMock(client, nil, nil, nil, cfg))
}
