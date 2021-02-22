package auditlog

import (
	"testing"

	"github.com/confluentinc/cli/internal/pkg/errors"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	climock "github.com/confluentinc/cli/mock"
)

func TestAuditLogDescribe(t *testing.T) {
	cmd := mockAuditLogCommand(true)

	_, err := pcmd.ExecuteCommand(cmd, "describe")
	require.NoError(t, err)
}

func TestAuditLogDescribeUnconfigured(t *testing.T) {
	cmd := mockAuditLogCommand(false)

	out, err := pcmd.ExecuteCommand(cmd, "describe")
	require.Error(t, err)
	require.Equal(t, "Error: "+errors.AuditLogsNotEnabledErrorMsg+"\n", out)
}

func mockAuditLogCommand(configured bool) *cobra.Command {
	client := &ccloud.Client{}
	cfg := v3.AuthenticatedCloudConfigMock()
	if configured {
		cfg.Context().State.Auth.Organization.AuditLog = &orgv1.AuditLog{
			ClusterId:        "lkc-ab123",
			AccountId:        "env-zy987",
			ServiceAccountId: 12345,
			TopicName:        "confluent-audit-log-events",
		}
	}
	return New("ccloud", climock.NewPreRunnerMock(client, nil, nil, cfg))
}
