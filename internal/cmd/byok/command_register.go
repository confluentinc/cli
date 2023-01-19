package byok

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"

	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/byok/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/aws/aws-sdk-go/aws/arn"
)

var encryptionKeyPolicyAWS = template.Must(template.New("encryptionKey").Parse(`{{range  $i, $accountID := .}}{{if $i}},{{end}}{
    "Sid" : "Allow Confluent accounts to use the key",
    "Effect" : "Allow",
    "Principal" : {
      "AWS" : ["{{$accountID}}"]
    },
    "Action" : [ "kms:Encrypt", "kms:Decrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey" ],
    "Resource" : "*"
  }, {
    "Sid" : "Allow Confluent accounts to attach persistent resources",
    "Effect" : "Allow",
    "Principal" : {
      "AWS" : ["{{$accountID}}"]
    },
    "Action" : [ "kms:CreateGrant", "kms:ListGrants", "kms:RevokeGrant" ],
    "Resource" : "*"
}{{end}}`))

var keyVaultCryptoServiceEncryptionUser = "e147488a-f6f5-4113-8e2d-b22465e65bf6" //TODO: to const somewhere
var keyVaultReader = "21090545-7ca7-4776-b22c-e363652d74d2"                      //TODO: to const somewhere

var (
	fields            = []string{"Id", "Key", "Roles", "Provider", "State", "CreatedAt", "UpdatedAt", "DeletedAt"}
	humanRenames      = map[string]string{"Id": "ID", "Key": "Key", "Roles": "Roles", "Provider": "Provider", "State": "State", "CreatedAt": "Created At", "UpdatedAt": "Updated At", "DeletedAt": "Deleted At"}
	structuredRenames = map[string]string{"Id": "id", "Key": "key", "Roles": "roles", "Provider": "provider", "State": "state", "CreatedAt": "created_at", "UpdatedAt": "updated_at", "DeletedAt": "deleted_at"}
)

func (c *command) newRegisterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "register <key>",
		Short: "Register a self-managed key in Confluent Cloud.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.register,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new self-managed encryption key for AWS:",
				Code: `confluent byok register "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"`,
			},
			examples.Example{
				Text: "Register a new self-managed encryption key for Azure:",
				Code: `confluent byok register "https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000" --tenant_id "00000000-0000-0000-0000-000000000000" --key_vault_id "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault"`,
			},
		),
	}

	cmd.Flags().String("key_vault_id", "", "The ID of the Azure Key Vault where the key is stored.")
	cmd.Flags().String("tenant_id", "", "The ID of the Azure Active Directory tenant that the key vault belongs to.")
	cmd.MarkFlagsRequiredTogether("key_vault_id", "tenant_id")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) registerAWS(cmd *cobra.Command, keyString string) (*byokv1.ByokV1Key, error) {
	keyReq := byokv1.ByokV1Key{}

	keyReq.Key = &byokv1.ByokV1KeyKeyOneOf{
		ByokV1AwsKey: &byokv1.ByokV1AwsKey{
			KeyArn: keyString,
			Kind:   "AwsKey",
		},
	}

	keyResp, _, err := c.V2Client.CreateByokKey(keyReq)
	if err != nil {
		return nil, err
	}

	return &keyResp, nil
}

func (c *command) registerAzure(cmd *cobra.Command, keyString string) (*byokv1.ByokV1Key, error) {
	keyReq := byokv1.ByokV1Key{}

	keyVaultID, err := cmd.Flags().GetString("key_vault_id")
	if err != nil {
		return nil, err
	}
	tenantID, err := cmd.Flags().GetString("tenant_id")
	if err != nil {
		return nil, err
	}

	keyReq.Key = &byokv1.ByokV1KeyKeyOneOf{
		ByokV1AzureKey: &byokv1.ByokV1AzureKey{
			KeyId:      keyString,
			KeyVaultId: keyVaultID,
			TenantId:   tenantID,
			Kind:       "AzureKey",
		},
	}

	keyResp, _, err := c.V2Client.CreateByokKey(keyReq)
	if err != nil {
		return nil, err
	}

	return &keyResp, err

}

func (c *command) register(cmd *cobra.Command, args []string) error {
	keyString := args[0]

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	var policyInstructions string
	var key *byokv1.ByokV1Key

	if cmd.Flags().Changed("key_vault_id") && cmd.Flags().Changed("tenant_id") {
		key, err = c.registerAzure(cmd, keyString)
		if err != nil {
			return err
		}

		policyInstructions, err = renderAzureEncryptionPolicy(key)
		if err != nil {
			return err
		}
	} else if isAWSKey(keyString) {
		key, err = c.registerAWS(cmd, keyString)
		if err != nil {
			return err
		}

		policyInstructions, err = renderAWSEncryptionPolicy(*key.Key.ByokV1AwsKey.Roles)
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("invalid key format: %s", keyString))
	}

	if outputFormat == output.Human.String() {
		utils.Printf(cmd, errors.CreatedResourceMsg, resource.ByokKey, *key.Id)
		utils.Printf(cmd, policyInstructions)
	} else {
		return output.StructuredOutput(outputFormat, policyInstructions) // TODO: output structured output
	}

	return nil
}

func isAWSKey(key string) bool {
	keyArn, err := arn.Parse(key)
	if err != nil {
		return false
	}

	return keyArn.Service == "kms" && keyArn.Resource[:4] == "key/"
}

func renderAWSEncryptionPolicy(roles []string) (string, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(errors.CopyBYOKAWSPermissionsHeaderMsg)
	buf.WriteString("\n\n")
	if err := encryptionKeyPolicyAWS.Execute(buf, roles); err != nil {
		return "", errors.New(errors.FailedToRenderKeyPolicyErrorMsg)
	}
	buf.WriteString("\n\n")
	return buf.String(), nil
}

func renderAzureEncryptionPolicy(key *byokv1.ByokV1Key) (string, error) {

	object_id := fmt.Sprintf("$(az ad sp show --id \"%s\" --query id --out tsv 2>/dev/null || az ad sp create --id \"%s\" --query id --out tsv)", *key.Key.ByokV1AzureKey.ApplicationId, *key.Key.ByokV1AzureKey.ApplicationId)

	regex := regexp.MustCompile(`^https://([^/.]+).vault.azure.net`)
	vaultName := regex.FindStringSubmatch(key.Key.ByokV1AzureKey.KeyId)[1]

	buf := new(bytes.Buffer)
	buf.WriteString(errors.RunBYOKAzurePermissionsHeaderMsg)
	buf.WriteString("\n\n")
	buf.WriteString("az role assignment create \\\n")
	buf.WriteString(fmt.Sprintf("    --role \"%s\" \\\n", keyVaultCryptoServiceEncryptionUser))
	buf.WriteString(fmt.Sprintf("    --scope \"$(az keyvault show --name \"%s\" --query id --output tsv)\" \\\n", vaultName))
	buf.WriteString(fmt.Sprintf("    --assignee-object-id \"%s\" \\\n", object_id))
	buf.WriteString("    --assignee-principal-type ServicePrincipal && \\\n")
	buf.WriteString("az role assignment create \\\n")
	buf.WriteString(fmt.Sprintf("    --role \"%s\" \\\n", keyVaultReader))
	buf.WriteString(fmt.Sprintf("    --scope \"$(az keyvault show --name \"%s\" --query id --output tsv)\" \\\n", vaultName))
	buf.WriteString(fmt.Sprintf("    --assignee-object-id \"%s\" \\\n", object_id))
	buf.WriteString("    --assignee-principal-type ServicePrincipal\n\n")
	buf.WriteString("\n\n")

	return buf.String(), nil
}
