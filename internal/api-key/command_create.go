package apikey

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

type createOut struct {
	ApiKey    string `human:"API Key" serialized:"api_key"`
	ApiSecret string `human:"API Secret" serialized:"api_secret"`
}

var resourceTypeToKind = map[string]string{
	resource.KafkaCluster:          "Cluster",
	resource.KsqlCluster:           "ksqlDB",
	resource.SchemaRegistryCluster: "SchemaRegistry",
	resource.Cloud:                 "Cloud",
}

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create API keys for a given resource.",
		Long:  `Create API keys for a given resource. A resource is some Confluent product or service for which an API key can be created, for example ksqlDB application ID, or "cloud" to create a Cloud API key.`,
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an API key with full access to cluster "lkc-123456":`,
				Code: "confluent api-key create --resource lkc-123456",
			},
			examples.Example{
				Text: `Create an API key for Kafka cluster "lkc-123456" and service account "sa-123456":`,
				Code: "confluent api-key create --resource lkc-123456 --service-account sa-123456",
			},
			examples.Example{
				Text: `Create an API key for Schema Registry cluster "lsrc-123456":`,
				Code: "confluent api-key create --resource lsrc-123456",
			},
			examples.Example{
				Text: `Create an API key for KSQL cluster "lksqlc-123456":`,
				Code: "confluent api-key create --resource lksqlc-123456",
			},
			examples.Example{
				Text: "Create a Cloud API key:",
				Code: "confluent api-key create --resource cloud",
			},
		),
	}

	c.addResourceFlag(cmd, true)
	cmd.Flags().String("description", "", "Description of API key.")
	cmd.Flags().Bool("use", false, "Use the created API key for the provided resource.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("resource"))

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}
	if serviceAccount == "" {
		serviceAccount = c.Context.GetCurrentServiceAccount()
	}

	owner := serviceAccount
	if owner == "" {
		userId, err := c.getCurrentUserId()
		if err != nil {
			return err
		}
		owner = userId
	}

	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.V2Client)
	if err != nil {
		return err
	}

	key := apikeysv2.IamV2ApiKey{Spec: &apikeysv2.IamV2ApiKeySpec{
		Description: apikeysv2.PtrString(description),
		Owner:       &apikeysv2.ObjectReference{Id: owner},
		Resource: &apikeysv2.ObjectReference{
			Id:   clusterId,
			Kind: apikeysv2.PtrString(resourceTypeToKind[resourceType]),
		},
	}}
	if resourceType == resource.Cloud {
		key.Spec.Resource.Id = "cloud"
	}

	v2Key, httpResp, err := c.V2Client.CreateApiKey(key)
	if err != nil {
		return c.catchServiceAccountNotValidError(err, httpResp, clusterId, serviceAccount)
	}

	userKey := &config.APIKeyPair{
		Key:    v2Key.GetId(),
		Secret: v2Key.Spec.GetSecret(),
	}

	if output.GetFormat(cmd) == output.Human {
		output.ErrPrintln("It may take a couple of minutes for the API key to be ready.")
		output.ErrPrintln("Save the API key and secret. The secret is not retrievable later.")
	}

	table := output.NewTable(cmd)
	table.Add(&createOut{
		ApiKey:    userKey.Key,
		ApiSecret: userKey.Secret,
	})
	if err := table.Print(); err != nil {
		return err
	}

	if resourceType == resource.KafkaCluster {
		if err := c.keystore.StoreAPIKey(userKey, clusterId); err != nil {
			return errors.Wrap(err, unableToStoreApiKeyErrorMsg)
		}
	}

	use, err := cmd.Flags().GetBool("use")
	if err != nil {
		return err
	}
	if use {
		if resourceType != resource.KafkaCluster {
			return errors.Wrap(errors.New(nonKafkaNotImplementedErrorMsg), "`--use` set but ineffective")
		}
		if err := c.Context.UseAPIKey(userKey.Key, clusterId); err != nil {
			return errors.NewWrapErrorWithSuggestions(err, apiKeyUseFailedErrorMsg, fmt.Sprintf(apiKeyUseFailedSuggestions, userKey.Key))
		}
		output.Printf(useAPIKeyMsg, userKey.Key)
	}

	return nil
}

func (c *command) getCurrentUserId() (string, error) {
	users, err := c.getAllUsers()
	if err != nil {
		return "", err
	}
	for _, user := range users {
		if user.GetId() == c.Context.GetUser().GetId() {
			return user.GetResourceId(), nil
		}
	}
	return "", fmt.Errorf("unable to find authenticated user")
}

// CLI-1544: Warn users if they try to create an API key with the predefined audit log Kafka cluster, but without the
// predefined audit log service account
func (c *command) catchServiceAccountNotValidError(err error, httpResp *http.Response, clusterId, serviceAccountId string) error {
	if err == nil {
		return nil
	}

	if err.Error() == "error creating api key: service account is not valid" || err.Error() == "403 Forbidden" {
		user, err := c.Client.Auth.User()
		if err != nil {
			return err
		}

		auditLog := user.GetOrganization().GetAuditLog()
		if clusterId == auditLog.GetClusterId() && serviceAccountId != auditLog.GetServiceAccountResourceId() {
			return fmt.Errorf(`API keys for audit logs (limit of 2) must be created using the predefined service account, "%s"`, auditLog.GetServiceAccountResourceId())
		}
	}

	if httpResp == nil {
		return err
	}

	return errors.CatchCCloudV2Error(err, httpResp)
}
