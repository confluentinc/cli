package byok

import (
	"bytes"
	"fmt"
	"html/template"
	"net/url"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/spf13/cobra"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

var encryptionKeyPolicyAws = template.Must(template.New("encryptionKeyPolicyAws").Parse(`{
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

const (
	failedToRenderKeyPolicyErrorMsg = "BYOK error: failed to render key policy"

	keyVaultCryptoServiceEncryptionUser = "e147488a-f6f5-4113-8e2d-b22465e65bf6"
	keyVaultReader                      = "21090545-7ca7-4776-b22c-e363652d74d2"

	defaultGcpRoleName = "custom_kms_role"
)

func (c *command) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <key>",
		Short: "Register a self-managed encryption key.",
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
				Code: `confluent byok create "https://vault-name.vault.azure.net/keys/key-name" --tenant "00000000-0000-0000-0000-000000000000" --key-vault "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/resourcegroup-name/providers/Microsoft.KeyVault/vaults/vault-name"`,
			},
			examples.Example{
				Text: "Register a new self-managed encryption key for GCP:",
				Code: `confluent byok create "projects/exampleproject/locations/us-central1/keyRings/testkeyring/cryptoKeys/testbyokkey/cryptoKeyVersions/3"`,
			},
		),
	}

	cmd.Flags().String("key-vault", "", "The ID of the Azure Key Vault where the key is stored.")
	cmd.Flags().String("tenant", "", "The ID of the Azure Active Directory tenant that the key vault belongs to.")
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsRequiredTogether("key-vault", "tenant")

	return cmd
}

func (c *command) createAwsKeyRequest(keyArn string) byokv1.ByokV1Key {
	return byokv1.ByokV1Key{Key: &byokv1.ByokV1KeyKeyOneOf{ByokV1AwsKey: &byokv1.ByokV1AwsKey{
		KeyArn: keyArn,
		Kind:   "AwsKey",
	}}}
}

func (c *command) createAzureKeyRequest(cmd *cobra.Command, keyString string) (byokv1.ByokV1Key, error) {
	keyVault, err := cmd.Flags().GetString("key-vault")
	if err != nil {
		return byokv1.ByokV1Key{}, err
	}

	tenant, err := cmd.Flags().GetString("tenant")
	if err != nil {
		return byokv1.ByokV1Key{}, err
	}

	keyReq := byokv1.ByokV1Key{Key: &byokv1.ByokV1KeyKeyOneOf{ByokV1AzureKey: &byokv1.ByokV1AzureKey{
		KeyId:      keyString,
		KeyVaultId: keyVault,
		TenantId:   tenant,
		Kind:       "AzureKey",
	}}}

	return keyReq, nil
}

func (c *command) createGcpKeyRequest(keyString string) byokv1.ByokV1Key {
	return byokv1.ByokV1Key{Key: &byokv1.ByokV1KeyKeyOneOf{ByokV1GcpKey: &byokv1.ByokV1GcpKey{
		KeyId: keyString,
		Kind:  "GcpKey",
	}}}
}

func (c *command) create(cmd *cobra.Command, args []string) error {
	keyString := args[0]
	var keyReq byokv1.ByokV1Key

	switch {
	case cmd.Flags().Changed("key-vault") && cmd.Flags().Changed("tenant"):
		keyString = removeKeyVersionFromAzureKeyId(keyString)

		request, err := c.createAzureKeyRequest(cmd, keyString)
		if err != nil {
			return err
		}
		keyReq = request
	case isAWSKey(keyString):
		keyReq = c.createAwsKeyRequest(keyString)
	case isGcpKey(keyString):
		keyReq = c.createGcpKeyRequest(keyString)
	default:
		return fmt.Errorf("invalid key format: %s", keyString)
	}

	key, httpResp, err := c.V2Client.CreateByokKey(keyReq)
	if err != nil {
		return errors.CatchCCloudV2Error(err, httpResp)
	}

	return c.outputByokKeyDescription(cmd, key)
}

func isAWSKey(key string) bool {
	keyArn, err := arn.Parse(key)
	if err != nil {
		return false
	}

	return keyArn.Service == "kms" && strings.HasPrefix(keyArn.Resource, "key/")
}

func isGcpKey(key string) bool {
	return strings.HasPrefix(key, "projects/")
}

func getPolicyCommand(key byokv1.ByokV1Key) (string, error) {
	switch {
	case key.Key.ByokV1AwsKey != nil:
		return renderAWSEncryptionPolicy(key.Key.ByokV1AwsKey.GetRoles())
	case key.Key.ByokV1AzureKey != nil:
		return renderAzureEncryptionPolicy(key)
	case key.Key.ByokV1GcpKey != nil:
		return renderGCPEncryptionPolicy(key), nil
	default:
		return "", nil
	}
}

func renderAWSEncryptionPolicy(roles []string) (string, error) {
	buf := new(bytes.Buffer)
	if err := encryptionKeyPolicyAws.Execute(buf, roles); err != nil {
		return "", fmt.Errorf(failedToRenderKeyPolicyErrorMsg)
	}
	return buf.String(), nil
}

func renderAzureEncryptionPolicy(key byokv1.ByokV1Key) (string, error) {
	objectId := fmt.Sprintf(`$(az ad sp show --id "%s" --query id --out tsv 2>/dev/null || az ad sp create --id "%s" --query id --out tsv)`, key.Key.ByokV1AzureKey.GetApplicationId(), key.Key.ByokV1AzureKey.GetApplicationId())

	regex := regexp.MustCompile(`^https://([^/.]+).vault.azure.net`)
	matches := regex.FindStringSubmatch(key.Key.ByokV1AzureKey.KeyId)
	if matches == nil {
		return "", fmt.Errorf(failedToRenderKeyPolicyErrorMsg)
	}

	vaultName := matches[1]

	az := []string{
		`az role assignment create \`,
		fmt.Sprintf(`    --role "%s" \`, keyVaultCryptoServiceEncryptionUser),
		fmt.Sprintf(`    --scope "$(az keyvault show --name "%s" --query id --output tsv)" \`, vaultName),
		fmt.Sprintf(`    --assignee-object-id "%s" \`, objectId),
		`    --assignee-principal-type ServicePrincipal && \`,
		`az role assignment create \`,
		fmt.Sprintf(`    --role "%s" \`, keyVaultReader),
		fmt.Sprintf(`    --scope "$(az keyvault show --name "%s" --query id --output tsv)" \`, vaultName),
		fmt.Sprintf(`    --assignee-object-id "%s" \`, objectId),
		"    --assignee-principal-type ServicePrincipal",
	}

	return strings.Join(az, "\n"), nil
}

func renderGCPEncryptionPolicy(key byokv1.ByokV1Key) string {
	// No need to do a sanity check for this key as it is handled by the BYOK API validation
	// assumed to be a valid key ID
	splitKeyId := strings.Split(key.Key.ByokV1GcpKey.KeyId, "/")
	project, location, keyRing, keyName := splitKeyId[1], splitKeyId[3], splitKeyId[5], splitKeyId[7]
	group := key.Key.ByokV1GcpKey.GetSecurityGroup()

	encryptionPolicyMetadata := gcpPolicyMetadata{
		project:  project,
		location: location,
		keyRing:  keyRing,
		key:      keyName,
		group:    group,
	}

	return encryptionPolicyMetadata.renderPolicy()
}

func getPostCreateStepInstruction(key byokv1.ByokV1Key) string {
	switch {
	case key.Key.ByokV1AwsKey != nil:
		return `Copy and append these permissions into the key policy "Statements" field of the ARN in your AWS key management system to authorize access for your Confluent Cloud cluster.`
	case key.Key.ByokV1AzureKey != nil:
		return "To ensure the key vault has the correct role assignments, please run the following Azure CLI command (certified for `az` v2.45):"
	case key.Key.ByokV1GcpKey != nil:
		return "To ensure the key has the correct role assignments, please run the following Google Cloud CLI command:"
	default:
		return ""
	}
}

// Best effort to remove the key version from the Azure Key ID if it is present
// For any errors, return the original key ID as is
// All further validation of the key ID is done by the BYOK API
func removeKeyVersionFromAzureKeyId(keyId string) string {
	path, err := url.Parse(keyId)
	if err != nil || len(strings.Split(path.Path, "/")) != 4 {
		return keyId
	}

	pathSegments := strings.Split(path.Path, "/")
	return keyId[:len(keyId)-len(pathSegments[3])-1]
}
