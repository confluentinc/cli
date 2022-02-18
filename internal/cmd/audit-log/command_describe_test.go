package auditlog

import (
	"context"
	"testing"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go-v1"
	ccloudmock "github.com/confluentinc/ccloud-sdk-go-v1/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	climock "github.com/confluentinc/cli/mock"
)

func TestAuditLogDescribe(t *testing.T) {
	cmd := mockAuditLogCommand(true)

	_, err := pcmd.ExecuteCommand(cmd, "describe")
	require.NoError(t, err)
}

func TestAuditLogDescribeUnconfigured(t *testing.T) {
	cmd := mockAuditLogCommand(false)

	_, err := pcmd.ExecuteCommand(cmd, "describe")
	require.Error(t, err)
	require.Equal(t, errors.AuditLogsNotEnabledErrorMsg, err.Error())
}

func mockAuditLogCommand(configured bool) *cobra.Command {
	client := &ccloud.Client{
		User: &ccloudmock.User{
			GetServiceAccountFunc: func(_ context.Context, id int32) (*orgv1.User, error) {
				return &orgv1.User{ResourceId: "sa-123456"}, nil
			},
		},
	}

	cfg := v1.AuthenticatedCloudConfigMock()

	if configured {
		cfg.Context().State.Auth.Organization.AuditLog = &orgv1.AuditLog{
			ClusterId:        "lkc-ab123",
			AccountId:        "env-zy987",
			ServiceAccountId: 12345,
			TopicName:        "confluent-audit-log-events",
		}
	}

	return New(climock.NewPreRunnerMock(client, nil, nil, nil, nil, cfg))
}
