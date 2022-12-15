package byok

import (
	"strings"

	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/byok/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"

	"github.com/aws/aws-sdk-go/aws/arn"
)

var (
	fields            = []string{"Id", "Key", "Provider", "State", "CreatedAt", "UpdatedAt", "DeletedAt"}
	humanRenames      = map[string]string{"Id": "ID", "Key": "Key", "Provider": "Provider", "State": "State", "CreatedAt": "Created At", "UpdatedAt": "Updated At", "DeletedAt": "Deleted At"}
	structuredRenames = map[string]string{"Id": "ID", "Key": "Key", "Provider": "Provider", "State": "State", "CreatedAt": "created_at", "UpdatedAt": "updated_at", "DeletedAt": "deleted_at"}
)

func (c *command) newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <key>",
		Short: "Register a customer managed key in Confluent Cloud.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.register,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) register(cmd *cobra.Command, args []string) error {
	keyString := args[0]

	keyReq := byokv1.ByokV1Key{}

	switch {
	case isAWSKey(keyString):
		keyReq.Key = &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AwsKey: &byokv1.ByokV1AwsKey{
				KeyArn: keyString,
				Kind:   "AwsKey",
			},
		}
	case isAzureKey(keyString):
		keyReq.Key = &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AzureKey: &byokv1.ByokV1AzureKey{
				KeyId: keyString,
				Kind:  "AzureKey",
			},
		}
	default:
		return errors.New("invalid key format")
	}

	keyResp, _, err := c.V2Client.CreateByokKey(keyReq)
	if err != nil {
		return err
	}

	describeByokKey := &byokKey{
		Id:        *keyResp.Id,
		Provider:  *keyResp.Provider,
		State:     *keyResp.State,
		CreatedAt: keyResp.Metadata.CreatedAt.String(),
	}

	switch {
	case keyResp.Key.ByokV1AwsKey != nil:
		describeByokKey.Key = keyResp.Key.ByokV1AwsKey.KeyArn
	case keyResp.Key.ByokV1AzureKey != nil:
		describeByokKey.Key = keyResp.Key.ByokV1AzureKey.KeyId
	}

	return output.DescribeObject(cmd, describeByokKey, fields, humanRenames, structuredRenames)
}

func isAWSKey(key string) bool {
	keyArn, err := arn.Parse(key)
	if err != nil {
		return false
	}

	if keyArn.Service != "kms" {
		return false
	}

	if keyArn.Resource[:4] != "key/" {
		return false
	}

	return true
}

func isAzureKey(key string) bool {
	return strings.Contains(key, "vault.azure.net/keys")
}
