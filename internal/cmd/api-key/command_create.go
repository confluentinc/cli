package apikey

import (
	"context"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	createFields            = []string{"Key", "Secret"}
	createHumanRenames      = map[string]string{"Key": "API Key"}
	createStructuredRenames = map[string]string{"Key": "key", "Secret": "secret"}
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create API keys for a given resource.",
		Long:  `Create API keys for a given resource. A resource is some Confluent product or service for which an API key can be created, for example ksqlDB application ID, or "cloud" to create a Cloud API key.`,
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an API key for service account "sa-123456" for cluster "lkc-123456".`,
				Code: "confluent api-key create --resource lkc-123456 --service-account sa-123456",
			},
		),
	}
	cmd.Flags().String(resourceFlagName, "", `The resource ID. Use "cloud" to create a Cloud API key.`)
	cmd.Flags().String("service-account", "", "Service account ID. If not specified, the API key will have full access on the cluster.")
	cmd.Flags().String("description", "", "Description of API key.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(resourceFlagName)

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Config.Resolver, c.Client)
	if err != nil {
		return err
	}
	serviceAccountID, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	key := &schedv1.ApiKey{
		UserResourceId: serviceAccountID,
		Description:    description,
		AccountId:      c.EnvironmentId(),
	}

	key, err = c.completeKeyUserId(key) // get corresponding numeric ID if the cmd has a service-account flag
	if err != nil {
		return err
	}

	if resourceType != pcmd.CloudResourceType {
		key.LogicalClusters = []*schedv1.ApiKey_Cluster{{Id: clusterId, Type: resourceType}}
	}
	userKey, err := c.Client.APIKey.Create(context.Background(), key)
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.ErrPrintln(cmd, errors.APIKeyTime)
		utils.ErrPrintln(cmd, errors.APIKeyNotRetrievableMsg)
	}

	err = output.DescribeObject(cmd, userKey, createFields, createHumanRenames, createStructuredRenames)
	if err != nil {
		return err
	}

	if resourceType == pcmd.KafkaResourceType {
		if err := c.keystore.StoreAPIKey(userKey, clusterId); err != nil {
			return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
		}
	}

	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, userKey.UserResourceId)
	c.analyticsClient.SetSpecialProperty(analytics.ApiKeyPropertiesKey, userKey.Key)
	return nil
}

func (c *command) completeKeyUserId(key *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	if key.UserResourceId != "" { // it has a service-account flag
		validFormat := strings.HasPrefix(key.UserResourceId, "sa-")
		if !validFormat {
			return nil, errors.New(errors.BadServiceAccountIDErrorMsg)
		}
		users, err := c.getAllUsers()
		if err != nil {
			return key, err
		}
		for _, user := range users {
			if key.UserResourceId == user.ResourceId {
				key.UserId = user.Id
				break
			}
		}
	} else {
		key.ServiceAccount = false
	}
	return key, nil
}
