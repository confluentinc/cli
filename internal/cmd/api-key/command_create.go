package apikey

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
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
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddServiceAccountFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(resourceFlagName)

	return cmd
}

func (c *command) create(cmd *cobra.Command, _ []string) error {
	c.setKeyStoreIfNil()
	resourceType, clusterId, _, err := c.resolveResourceId(cmd, c.Client)
	if err != nil {
		return err
	}

	ownerResourceId, err := cmd.Flags().GetString("service-account")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	var userKey *v1.APIKeyPair
	if resourceType == resource.Ksql || resourceType == resource.SchemaRegistry {
		userKey, err = c.createV1(ownerResourceId, clusterId, resourceType, description)
		if err != nil {
			return err
		}
	} else {
		ownerResourceId, err = c.getApiKeyOwnerId(ownerResourceId)
		if err != nil {
			return err
		}

		key := apikeysv2.IamV2ApiKey{
			Spec: &apikeysv2.IamV2ApiKeySpec{
				Description: apikeysv2.PtrString(description),
				Owner:       &apikeysv2.ObjectReference{Id: ownerResourceId},
				Resource: &apikeysv2.ObjectReference{
					Id:   clusterId,
					Kind: apikeysv2.PtrString(resourceTypeToKind[resourceType]),
				},
			},
		}
		if resourceType == resource.Cloud {
			key.Spec.Resource.Id = "cloud"
		}

		v2Key, _, err := c.V2Client.CreateApiKey(key)
		if err != nil {
			return c.catchServiceAccountNotValidError(err, clusterId, ownerResourceId)
		}

		userKey = &v1.APIKeyPair{
			Key:    *v2Key.Id,
			Secret: *v2Key.Spec.Secret,
		}
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

	if resourceType == resource.Kafka {
		if err := c.keystore.StoreAPIKey(userKey, clusterId); err != nil {
			return errors.Wrap(err, errors.UnableToStoreAPIKeyErrorMsg)
		}
	}

	return nil
}

func (c *command) createV1(ownerResourceId, clusterId, resourceType, description string) (*v1.APIKeyPair, error) {
	key := &schedv1.ApiKey{
		UserResourceId: ownerResourceId,
		Description:    description,
		AccountId:      c.EnvironmentId(),
	}

	key, err := c.completeKeyUserId(key) // get corresponding numeric ID if the cmd has a service-account flag
	if err != nil {
		return nil, err
	}

	if resourceType != resource.Cloud {
		key.LogicalClusters = []*schedv1.ApiKey_Cluster{{Id: clusterId, Type: resourceType}}
	}
	schedv1ApiKey, err := c.Client.APIKey.Create(context.Background(), key)
	displayKey := &v1.APIKeyPair{
		Key:    schedv1ApiKey.Key,
		Secret: schedv1ApiKey.Secret,
	}
	return displayKey, c.catchServiceAccountNotValidError(err, clusterId, ownerResourceId)

}

func (c *command) completeKeyUserId(key *schedv1.ApiKey) (*schedv1.ApiKey, error) {
	if key.UserResourceId != "" { // it has a service-account flag
		if resource.LookupType(key.UserResourceId) != resource.ServiceAccount {
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

func (c *command) getApiKeyOwnerId(ownerResourceId string) (string, error) {
	if ownerResourceId != "" {
		return ownerResourceId, nil
	}

	users, err := c.getAllUsers()
	if err != nil {
		return "", err
	}
	for _, user := range users {
		if user.Id == c.State.Auth.User.Id {
			return user.ResourceId, nil
		}
	}
	return ownerResourceId, nil
}

// CLI-1544: Warn users if they try to create an API key with the predefined audit log Kafka cluster, but without the
// predefined audit log service account
func (c *command) catchServiceAccountNotValidError(err error, clusterId, serviceAccountId string) error {
	if err == nil {
		return nil
	}
	isInvalid := err.Error() == "error creating api key: service account is not valid" || err.Error() == "403 Forbidden"
	if isInvalid && clusterId == c.State.Auth.Organization.AuditLog.ClusterId {
		auditLogServiceAccount, err2 := c.Client.User.GetServiceAccount(context.Background(), c.State.Auth.Organization.AuditLog.ServiceAccountId)
		if err2 != nil {
			return err
		}

		if serviceAccountId != auditLogServiceAccount.ResourceId {
			return fmt.Errorf(`API keys for audit logs (limit of 2) must be created using the predefined service account, "%s"`, auditLogServiceAccount.ResourceId)
		}
	}

	return err
}
