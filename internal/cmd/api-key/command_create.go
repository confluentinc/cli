package apikey

import (
	"context"
	"fmt"
	"net/http"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
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
				Text: `Create an API key for cluster "lkc-123456" and service account "sa-123456":`,
				Code: "confluent api-key create --resource lkc-123456 --service-account sa-123456",
			},
		),
	}

	cmd.Flags().String(resourceFlagName, "", `The resource ID. Use "cloud" to create a Cloud API key.`)
	cmd.Flags().String("description", "", "Description of API key.")
	cmd.Flags().Bool("use", false, "Use the created apikey for the provided resource.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(resourceFlagName)

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.V2Client)
	if err != nil {
		return err
	}

	serviceAccount, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	if serviceAccount == "" {
		serviceAccount, err = c.getCurrentUserId()
		if err != nil {
			return err
		}
	}

	key := apikeysv2.IamV2ApiKey{
		Spec: &apikeysv2.IamV2ApiKeySpec{
			Description: apikeysv2.PtrString(description),
			Owner:       &apikeysv2.ObjectReference{Id: serviceAccount},
			Resource: &apikeysv2.ObjectReference{
				Id:   clusterId,
				Kind: apikeysv2.PtrString(resourceTypeToKind[resourceType]),
			},
		},
	}
	if resourceType == resource.Cloud {
		key.Spec.Resource.Id = "cloud"
	}

	v2Key, httpResp, err := c.V2Client.CreateApiKey(key)
	if err != nil {
		return c.catchServiceAccountNotValidError(err, httpResp, clusterId, serviceAccount)
	}

	userKey := &v1.APIKeyPair{
		Key:    v2Key.GetId(),
		Secret: v2Key.Spec.GetSecret(),
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.ErrPrintln(cmd, errors.APIKeyTime)
		utils.ErrPrintln(cmd, errors.APIKeyNotRetrievableMsg)
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
			return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
		}
	}

	use, err := cmd.Flags().GetBool("use")
	if err != nil {
		return err
	}
	if use {
		if resourceType != resource.KafkaCluster {
			return errors.Wrap(errors.New(errors.NonKafkaNotImplementedErrorMsg), `"--use" set but ineffective`)
		}
		err = c.Context.UseAPIKey(userKey.Key, clusterId)
		if err != nil {
			return errors.NewWrapErrorWithSuggestions(err, errors.APIKeyUseFailedErrorMsg, fmt.Sprintf(errors.APIKeyUseFailedSuggestions, userKey.Key))
		}
		utils.Printf(cmd, errors.UseAPIKeyMsg, userKey.Key, clusterId)
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
func (c *command) catchServiceAccountNotValidError(err error, r *http.Response, clusterId, serviceAccountId string) error {
	if err == nil {
		return nil
	}

	auditLog := c.Context.GetOrganization().GetAuditLog()

	isInvalid := err.Error() == "error creating api key: service account is not valid" || err.Error() == "403 Forbidden"
	if isInvalid && clusterId == auditLog.GetClusterId() {
		auditLogServiceAccount, err2 := c.Client.User.GetServiceAccount(context.Background(), auditLog.GetServiceAccountId())
		if err2 != nil {
			return err
		}

		if serviceAccountId != auditLogServiceAccount.GetResourceId() {
			return fmt.Errorf(`API keys for audit logs (limit of 2) must be created using the predefined service account, "%s"`, auditLogServiceAccount.GetResourceId())
		}
	}

	if r == nil {
		return err
	}

	return errors.CatchCCloudV2Error(err, r)
}
