package byok

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"

	"github.com/aws/aws-sdk-go/aws/arn"
)

var encryptionKeyPolicyAWS = template.Must(template.New("encryptionKeyPolicyAWS").Parse(`{
	"Sid" : "Allow Confluent accounts to use the key",
	"Effect" : "Allow",
	"Principal" : {
		"AWS" : [{{range $i, $e := .}}{{if $i}},{{end}}
			"{{$e}}"{{end}}
		]
	},
	"Action" : [ "kms:Encrypt", "kms:Decrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey" ],
	"Resource" : "*"
}, {
	"Sid" : "Allow Confluent accounts to attach persistent resources",
	"Effect" : "Allow",
	"Principal" : {
		"AWS" : [{{range $i, $e := .}}{{if $i}},{{end}}
			"{{$e}}"{{end}}
		]
	},
	"Action" : [ "kms:CreateGrant", "kms:ListGrants", "kms:RevokeGrant" ],
	"Resource" : "*"
}`))

const keyVaultCryptoServiceEncryptionUser = "e147488a-f6f5-4113-8e2d-b22465e65bf6"
const keyVaultReader = "21090545-7ca7-4776-b22c-e363652d74d2"

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <key>",
		Short: "Register a self-managed encryption key",
		Long:  "Bring your own key to Confluent Cloud for data at rest encryption.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Register a new self-managed encryption key for AWS:",
				Code: `confluent byok create "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"`,
			},
			examples.Example{
				Text: "Register a new self-managed encryption key for Azure:",
				Code: `confluent byok create "https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000" --tenant "00000000-0000-0000-0000-000000000000" --key-vault "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault"`,
			},
		),
	}

	cmd.Flags().String("key-vault", "", "The ID of the Azure Key Vault where the key is stored.")
	cmd.Flags().String("tenant", "", "The ID of the Azure Active Directory tenant that the key vault belongs to.")
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("key-vault", "tenant")

	return cmd
}

func (c *command) createAwsKeyRequest(cmd *cobra.Command, keyString string) (*byokv1.ByokV1Key, error) {
	keyReq := byokv1.ByokV1Key{
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AwsKey: &byokv1.ByokV1AwsKey{
				KeyArn: keyString,
				Kind:   "AwsKey",
			},
		},
	}

	return &keyReq, nil
}

func (c *command) createAzureKeyRequest(cmd *cobra.Command, keyString string) (*byokv1.ByokV1Key, error) {
	keyVaultID, err := cmd.Flags().GetString("key-vault")
	if err != nil {
		return nil, err
	}
	tenantID, err := cmd.Flags().GetString("tenant")
	if err != nil {
		return nil, err
	}

	keyReq := byokv1.ByokV1Key{
		Key: &byokv1.ByokV1KeyKeyOneOf{
			ByokV1AzureKey: &byokv1.ByokV1AzureKey{
				KeyId:      keyString,
				KeyVaultId: keyVaultID,
				TenantId:   tenantID,
				Kind:       "AzureKey",
			},
		},
	}

	return &keyReq, err
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	keyString := args[0]
	var err error
	var keyReq *byokv1.ByokV1Key

	if cmd.Flags().Changed("key-vault") && cmd.Flags().Changed("tenant") {
		keyReq, err = c.createAzureKeyRequest(cmd, keyString)
		if err != nil {
			return err
		}
	} else if isAWSKey(keyString) {
		keyReq, err = c.createAwsKeyRequest(cmd, keyString)
		if err != nil {
			return err
		}
	} else {
		return errors.New(fmt.Sprintf("invalid key format: %s", keyString))
	}

	key, httpResp, err := c.V2Client.CreateByokKey(*keyReq)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	postCreationStepInstructions, err := getPostCreationStepInstructions(&key)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		utils.Printf(cmd, errors.CreatedResourceMsg, resource.ByokKey, key.GetId())
	}

	utils.Printf(cmd, postCreationStepInstructions)

	return nil
}

func isAWSKey(key string) bool {
	keyArn, err := arn.Parse(key)
	if err != nil {
		return false
	}

	return keyArn.Service == "kms" && strings.HasPrefix(keyArn.Resource, "key/")
}

func getPostCreationStepInstructions(key *byokv1.ByokV1Key) (string, error) {
	switch {
	case key.Key.ByokV1AwsKey != nil:
		return renderAWSEncryptionPolicy(*key.Key.ByokV1AwsKey.Roles)
	case key.Key.ByokV1AzureKey != nil:
		return renderAzureEncryptionPolicy(key)
	default:
		return "", nil
	}
}

func renderAWSEncryptionPolicy(roles []string) (string, error) {
	buf := new(bytes.Buffer)
	buf.WriteString(errors.CopyByokAwsPermissionsHeaderMsg)
	buf.WriteString("\n\n")
	if err := encryptionKeyPolicyAWS.Execute(buf, roles); err != nil {
		return "", errors.New(errors.FailedToRenderKeyPolicyErrorMsg)
	}
	buf.WriteString("\n\n")
	return buf.String(), nil
}

func renderAzureEncryptionPolicy(key *byokv1.ByokV1Key) (string, error) {
	objectId := fmt.Sprintf(`$(az ad sp show --id "%s" --query id --out tsv 2>/dev/null || az ad sp create --id "%s" --query id --out tsv)`, *key.Key.ByokV1AzureKey.ApplicationId, *key.Key.ByokV1AzureKey.ApplicationId)

	regex := regexp.MustCompile(`^https://([^/.]+).vault.azure.net`)
	vaultName := regex.FindStringSubmatch(key.Key.ByokV1AzureKey.KeyId)[1]

	az := []string{
		errors.RunByokAzurePermissionsHeaderMsg,
		"\n",
		"az role assignment create \\",
		fmt.Sprintf("    --role \"%s\" \\", keyVaultCryptoServiceEncryptionUser),
		fmt.Sprintf("    --scope \"$(az keyvault show --name \"%s\" --query id --output tsv)\" \\", vaultName),
		fmt.Sprintf("    --assignee-object-id \"%s\" \\", objectId),
		"    --assignee-principal-type ServicePrincipal && \\",
		"az role assignment create \\",
		fmt.Sprintf("    --role \"%s\" \\", keyVaultReader),
		fmt.Sprintf("    --scope \"$(az keyvault show --name \"%s\" --query id --output tsv)\" \\", vaultName),
		fmt.Sprintf("    --assignee-object-id \"%s\" \\", objectId),
		"    --assignee-principal-type ServicePrincipal",
	}

	return strings.Join(az, "\n"), nil
}
