package kafka

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/c-bata/go-prompt"
	corev1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	v1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/google/go-cmp/cmp"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	ccsdkmock "github.com/confluentinc/ccloud-sdk-go/mock"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	configv1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/mock"
	cliMock "github.com/confluentinc/cli/mock"
)

const (
	clusterId   = "lkc-0000"
	clusterName = "testCluster"
)

type KafkaClusterTestSuite struct {
	suite.Suite
}

func (suite *KafkaClusterTestSuite) newCmd(conf *v3.Config) *clusterCommand {
	client := &ccloud.Client{
		Kafka: &ccsdkmock.Kafka{
			ListFunc: func(_ context.Context, cluster *v1.KafkaCluster) ([]*v1.KafkaCluster, error) {
				return []*v1.KafkaCluster{
					{
						Id:   clusterId,
						Name: clusterName,
					},
				}, nil
			},
		},
	}
	prerunner := cliMock.NewPreRunnerMock(client, nil, conf)
	cmd := NewClusterCommand(prerunner)
	return cmd
}

func (suite *KafkaClusterTestSuite) TestServerComplete() {
	req := suite.Require()
	type fields struct {
		Command *clusterCommand
	}
	tests := []struct {
		name   string
		fields fields
		want   []prompt.Suggest
	}{
		{
			name: "suggest for authenticated user",
			fields: fields{
				Command: suite.newCmd(v3.AuthenticatedCloudConfigMock()),
			},
			want: []prompt.Suggest{
				{
					Text:        clusterId,
					Description: clusterName,
				},
			},
		},
		{
			name: "don't suggest for unauthenticated user",
			fields: fields{
				suite.newCmd(v3.UnauthenticatedCloudConfigMock()),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			got := tt.fields.Command.ServerComplete()
			fmt.Println(&got)
			req.Equal(tt.want, got)
		})
	}
}

func (suite *KafkaClusterTestSuite) TestCreateGCPBYOK() {
	req := require.New(suite.T())
	root := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	kafkaMock := &ccsdkmock.Kafka{
		CreateFunc: func(ctx context.Context, config *v1.KafkaClusterConfig) (*v1.KafkaCluster, error) {
			return &v1.KafkaCluster{
				Id:              "lkc-xyz",
				Name:            "gcp-byok-test",
				Region:          "us-central1",
				ServiceProvider: "gcp",
				Deployment: &v1.Deployment{
					Sku: corev1.Sku_DEDICATED,
				},
			}, nil
		},
	}
	idMock := &ccsdkmock.ExternalIdentity{
		CreateExternalIdentityFunc: func(_ context.Context, cloud, accountID string) (string, error) {
			return "id-xyz", nil
		},
	}
	client := &ccloud.Client{
		Kafka:            kafkaMock,
		ExternalIdentity: idMock,
		EnvironmentMetadata: &ccsdkmock.EnvironmentMetadata{
			GetFunc: func(ctx context.Context) ([]*schedv1.CloudMetadata, error) {
				return []*schedv1.CloudMetadata{{
					Id:       "gcp",
					Accounts: []*schedv1.AccountMetadata{{Id: "account-xyz"}},
					Regions:  []*schedv1.Region{{IsSchedulable: true, Id: "us-central1"}},
				}}, nil
			},
		},
	}
	root.AuthenticatedCLICommand.State = &v2.ContextState{
		Auth: &configv1.AuthConfig{
			Account: &orgv1.Account{
				Id: "abc",
			},
		},
	}
	root.Client = client
	var buf bytes.Buffer
	root.SetOut(&buf)
	cmd, args, err := root.Command.Find([]string{
		"create",
		"gcp-byok-test",
	})
	req.NoError(err)
	err = cmd.ParseFlags([]string{
		"--cloud=gcp",
		"--region=us-central1",
		"--type=dedicated",
		"--cku=1",
		"--encryption-key=xyz",
	})
	req.NoError(err)
	err = root.create(cmd, args, mock.NewPromptMock(
		"y", // yes customer has granted key access
	))
	req.NoError(err)
	got, want := buf.Bytes(), []byte(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  id-xyz


Please confirm you've authorized the key for this identity: id-xyz (y/n): It may take up to 5 minutes for the Kafka cluster to be ready.
`)
	req.True(cmp.Equal(got, want), cmp.Diff(got, want))
	req.Equal("abc", idMock.CreateExternalIdentityCalls()[0].AccountID)
	req.Equal("gcp", idMock.CreateExternalIdentityCalls()[0].Cloud)
	req.Equal("abc", kafkaMock.CreateCalls()[0].Config.AccountId)
	req.Equal("gcp", kafkaMock.CreateCalls()[0].Config.ServiceProvider)
	req.Equal("us-central1", kafkaMock.CreateCalls()[0].Config.Region)
	req.Equal("xyz", kafkaMock.CreateCalls()[0].Config.EncryptionKeyId)
	req.Equal(int32(1), kafkaMock.CreateCalls()[0].Config.Cku)
	req.Equal(corev1.Sku_DEDICATED, kafkaMock.CreateCalls()[0].Config.Deployment.Sku)
}

func (suite *KafkaClusterTestSuite) TestServerCompletableChildren() {
	req := require.New(suite.T())
	cmd := suite.newCmd(v3.AuthenticatedCloudConfigMock())
	completableChildren := cmd.ServerCompletableChildren()
	expectedChildren := []string{"cluster delete", "cluster describe", "cluster update", "cluster use"}
	req.Len(completableChildren, len(expectedChildren))
	for i, expectedChild := range expectedChildren {
		req.Contains(completableChildren[i].CommandPath(), expectedChild)
	}
}

func TestKafkaClusterTestSuite(t *testing.T) {
	suite.Run(t, new(KafkaClusterTestSuite))
}
