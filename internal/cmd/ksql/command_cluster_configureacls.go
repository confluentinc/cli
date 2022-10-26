package ksql

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	"github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *ksqlCommand) newConfigureAclsCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "configure-acls <id> TOPICS...",
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

	// Ensure the KSQL cluster talks to the current Kafka Cluster
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return err
	}
	if cluster.KafkaClusterId != kafkaCluster.Id {
		utils.ErrPrintf(cmd, errors.KsqlDBNotBackedByKafkaMsg, args[0], cluster.KafkaClusterId, kafkaCluster.Id, cluster.KafkaClusterId)
	}

	if cluster.ServiceAccountId == 0 {
		return fmt.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, args[0])
	}

	serviceAccountId, err := c.getServiceAccount(cluster)
	if err != nil {
		return err
	}

	// Setup ACLs
	aclsDryRun, err := cmd.Flags().GetBool("dry-run")
	if err != nil {
		return err
	}

	bindings := buildACLBindings(serviceAccountId, cluster, args[1:])
	if aclsDryRun {
		return acl.PrintACLs(cmd, bindings, cmd.OutOrStderr())
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
			if httpResp.StatusCode != http.StatusCreated {
				msg := fmt.Sprintf(errors.KafkaRestUnexpectedStatusErrorMsg, httpResp.Request.URL, httpResp.StatusCode)
				return errors.NewErrorWithSuggestions(msg, errors.InternalServerErrorSuggestions)
			}
			return nil
		}
	}

	return c.Client.Kafka.CreateACLs(context.Background(), kafkaCluster, bindings)
}

func (c *ksqlCommand) getServiceAccount(cluster *schedv1.KSQLCluster) (string, error) {
	users, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return "", err
	}

	for _, user := range users {
		if user.ServiceName == fmt.Sprintf("KSQL.%s", cluster.Id) || (cluster.KafkaApiKey != nil && user.Id == cluster.KafkaApiKey.UserId) {
			return strconv.Itoa(int(user.Id)), nil
		}
	}
	return "", errors.Errorf(errors.KsqlDBNoServiceAccountErrorMsg, cluster.Id)
}

func buildACLBindings(serviceAccountId string, cluster *schedv1.KSQLCluster, topics []string) []*schedv1.ACLBinding {
	var bindings []*schedv1.ACLBinding

	for _, operation := range []schedv1.ACLOperations_ACLOperation{
		schedv1.ACLOperations_DESCRIBE,
		schedv1.ACLOperations_DESCRIBE_CONFIGS,
	} {
		bindings = append(bindings, createACL(schedv1.ResourceTypes_CLUSTER, "kafka-cluster", schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
	}

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
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TOPIC, "_confluent-ksql-"+cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
		bindings = append(bindings, createACL(schedv1.ResourceTypes_GROUP, "_confluent-ksql-"+cluster.OutputTopicPrefix, schedv1.PatternTypes_PREFIXED, serviceAccountId, operation))
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
		bindings = append(bindings, createACL(schedv1.ResourceTypes_TRANSACTIONAL_ID, cluster.PhysicalClusterId, schedv1.PatternTypes_LITERAL, serviceAccountId, operation))
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
