package ksql

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newConfigureAclsCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "configure-acls <id> [topic-1] [topic-2] ... [topic-N]",
		Short:             fmt.Sprintf("Configure ACLs for a ksqlDB %s.", resource),
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.configureACLs,
	}

	cmd.Flags().Bool("dry-run", false, "If specified, print the ACLs that will be set and exit.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *ksqlCommand) configureACLs(cmd *cobra.Command, args []string) error {
	// Get the Kafka Cluster
	kafkaCluster, err := dynamicconfig.KafkaCluster(c.Context)
	if err != nil {
		return err
	}

	ksqlCluster := args[0]

	// Ensure the KSQL cluster talks to the current Kafka Cluster
	cluster, err := c.V2Client.DescribeKsqlCluster(ksqlCluster, c.EnvironmentId())
	if err != nil {
		return err
	}

	if cluster.Spec.KafkaCluster.Id != kafkaCluster.Id {
		utils.ErrPrintf(cmd, errors.KsqlDBNotBackedByKafkaMsg, ksqlCluster, cluster.Spec.KafkaCluster.Id, kafkaCluster.Id, cluster.Spec.KafkaCluster.Id)
	}

	credentialIdentity := cluster.Spec.CredentialIdentity.GetId()
	if resource.LookupType(credentialIdentity) != resource.ServiceAccount {
		return fmt.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, ksqlCluster)
	}

	serviceAccountId, err := c.getServiceAccount(&cluster)
	if err != nil {
		return err
	}

	// Setup ACLs
	aclsDryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	bindings := buildACLBindings(serviceAccountId, &cluster, args[1:])
	if aclsDryRun {
		return acl.PrintACLs(cmd, bindings)
	}

	if kafkaREST, _ := c.GetKafkaREST(); kafkaREST != nil {
		kafkaClusterConfig, err := c.Context.GetKafkaClusterForCommand()
		if err != nil {
			return err
		}

		httpResp, err := kafkaREST.CloudClient.BatchCreateKafkaAcls(kafkaClusterConfig.ID, getCreateAclRequestDataList(bindings))
		if err != nil && httpResp != nil {
			return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
		}
		if err == nil && httpResp != nil {
			if httpResp.StatusCode != http.StatusNoContent {
				msg := fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode)
				return errors.NewErrorWithSuggestions(msg, errors.InternalServerErrorSuggestions)
			}
			return nil
		}
	}

	return c.Client.Kafka.CreateACLs(context.Background(), kafkaCluster, bindings)
}

func (c *ksqlCommand) getServiceAccount(cluster *ksqlv2.KsqldbcmV2Cluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	credentialIdentity := cluster.Spec.CredentialIdentity.GetId()

	for _, user := range users {
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.GetId()) || user.ResourceId == credentialIdentity {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, cluster.GetId())
}

func buildACLBindings(serviceAccountId string, cluster *ksqlv2.KsqldbcmV2Cluster, topics []string) []*schedv1.ACLBinding {
	var bindings []*schedv1.ACLBinding

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, createACL(schedv1.ResourceTypes_CLUSTER, "kafka-cluster", schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	topicPrefix := cluster.Status.GetTopicPrefix()

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_CREATE,
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_ALTER,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
		schedv1.ACLOperations_ALTER_CONFIGS,
		schedv1.ACLOperations_READ,
		schedv1.ACLOperations_WRITE,
		schedv1.ACLOperations_DELETE,
	} {
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, topicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, "_confluent-ksql-"+topicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(schedv1.ResourceTypes_GROUP, "_confluent-ksql-"+topicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
	}

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, "*", schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
		bindings = append(bindings, createACL(schedv1.ResourceTypes_GROUP, "*", schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
		schedv1.ACLOperations_READ,
	} {
		for _, topic := range topics {
			bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, topic, schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
		}
	}

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_WRITE,
	} {
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TRANSACTIONAL_ID, topicPrefix, schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
	}

	return bindings
}

func createACL(resourceType schedv1.ResourceTypes_ResourceType, name string, patternType schedv1.PatternTypes_PatternType, serviceAccountId string, operation schedv1.ACLOperations_ACLOperation) *schedv1.ACLBinding {
	return &schedv1.ACLBinding{
		Pattern: &schedv1.ResourcePatternConfig{
			ResourceType: resourceType,
			Name:         name,
			PatternType:  patternType,
		},
		Entry: &schedv1.AccessControlEntryConfig{
			Principal:      "User:" + serviceAccountId,
			Operation:      operation,
			Host:           "*",
			PermissionType: schedv1.ACLPermissionTypes_ALLOW,
		},
	}
}

func getCreateAclRequestDataList(bindings []*schedv1.ACLBinding) kafkarestv3.CreateAclRequestDataList {
	data := make([]kafkarestv3.CreateAclRequestData, len(bindings))
	for i, binding := range bindings {
		data[i] = acl.GetCreateAclRequestData(binding)
	}
	return kafkarestv3.CreateAclRequestDataList{Data: data}
}
