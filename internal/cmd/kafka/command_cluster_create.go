package kafka

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"text/template"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	skuBasic     = "basic"
	skuStandard  = "standard"
	skuDedicated = "dedicated"
)

var encryptionKeyPolicy = template.Must(template.New("encryptionKey").Parse(`{{range  $i, $accountID := .}}{{if $i}},{{end}}{
    "Sid" : "Allow Confluent account ({{$accountID}}) to use the key",
    "Effect" : "Allow",
    "Principal" : {
      "AWS" : ["arn:aws:iam::{{$accountID}}:root"]
    },
    "Action" : [ "kms:Encrypt", "kms:Decrypt", "kms:ReEncrypt*", "kms:GenerateDataKey*", "kms:DescribeKey" ],
    "Resource" : "*"
  }, {
    "Sid" : "Allow Confluent account ({{$accountID}}) to attach persistent resources",
    "Effect" : "Allow",
    "Principal" : {
      "AWS" : ["arn:aws:iam::{{$accountID}}:root"]
    },
    "Action" : [ "kms:CreateGrant", "kms:ListGrants", "kms:RevokeGrant" ],
    "Resource" : "*"
}{{end}}`))

var permitBYOKGCP = template.Must(template.New("byok_gcp_permissions").Parse(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  {{.ExternalIdentity}}`))

type validateEncryptionKeyInput struct {
	Cloud          string
	MetadataClouds []*ccloudv1.CloudMetadata
	AccountID      string
}

func (c *clusterCommand) newCreateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Kafka cluster.",
		Long:  "Create a Kafka cluster.\n\nNote: You cannot use this command to create a cluster that is configured with AWS PrivateLink. You must use the UI to create a cluster of that configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.create(cmd, args, form.NewPrompt(os.Stdin))
		},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new dedicated cluster that uses a customer-managed encryption key in AWS:",
				Code: `confluent kafka cluster create sales092020 --cloud aws --region us-west-2 --type dedicated --cku 1 --encryption-key "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"`,
			},
			examples.Example{
				Text: "For more information, see https://docs.confluent.io/current/cloud/clusters/byok-encrypted-clusters.html.",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddAvailabilityFlag(cmd)
	pcmd.AddTypeFlag(cmd)
	cmd.Flags().Int("cku", 0, `Number of Confluent Kafka Units (non-negative). Required for Kafka clusters of type "dedicated".`)
	cmd.Flags().String("encryption-key", "", "Encryption Key ID (e.g. for Amazon Web Services, the Amazon Resource Name of the key).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("cloud")
	_ = cmd.MarkFlagRequired("region")

	return cmd
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string, prompt form.Prompt) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	clouds, err := c.Client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return err
	}

	if err := checkCloudAndRegion(cloud, region, clouds); err != nil {
		return err
	}

	availabilityString, err := cmd.Flags().GetString("availability")
	if err != nil {
		return err
	}

	availability, err := stringToAvailability(availabilityString)
	if err != nil {
		return err
	}

	clusterType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}

	sku, err := stringToSku(clusterType)
	if err != nil {
		return err
	}

	encryptionKey, err := cmd.Flags().GetString("encryption-key")
	if err != nil {
		return err
	}

	if encryptionKey != "" {
		input := validateEncryptionKeyInput{
			Cloud:          cloud,
			MetadataClouds: clouds,
			AccountID:      c.EnvironmentId(),
		}
		if err := c.validateEncryptionKey(cmd, prompt, input); err != nil {
			return err
		}
	}

	createCluster := cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			Environment: &cmkv2.ObjectReference{
				Id: c.EnvironmentId(),
			},
			DisplayName:  cmkv2.PtrString(args[0]),
			Cloud:        cmkv2.PtrString(cloud),
			Region:       cmkv2.PtrString(region),
			Availability: cmkv2.PtrString(availability),
			Config:       setCmkClusterConfig(clusterType, 1, encryptionKey),
		},
	}

	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetInt("cku")
		if err != nil {
			return err
		}
		if clusterType != skuDedicated {
			return errors.NewErrorWithSuggestions("the `--cku` flag can only be used when creating a dedicated Kafka cluster", "Specify a dedicated cluster with `--type`.")
		}
		if cku <= 0 {
			return errors.New(errors.CKUMoreThanZeroErrorMsg)
		}
		setClusterConfigCku(&createCluster, int32(cku))
	}

	kafkaCluster, httpResp, err := c.V2Client.CreateKafkaCluster(createCluster)
	if err != nil {
		return errors.CatchClusterConfigurationNotValidError(err, httpResp)
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.ErrPrintln(cmd, getKafkaProvisionEstimate(sku))
	}

	return c.outputKafkaClusterDescription(cmd, &kafkaCluster, false)
}

func checkCloudAndRegion(cloudId string, regionId string, clouds []*ccloudv1.CloudMetadata) error {
	for _, cloud := range clouds {
		if cloudId == cloud.GetId() {
			for _, region := range cloud.GetRegions() {
				if regionId == region.GetId() {
					if region.GetIsSchedulable() {
						return nil
					} else {
						break
					}
				}
			}
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.CloudRegionNotAvailableErrorMsg, regionId, cloudId),
				fmt.Sprintf(errors.CloudRegionNotAvailableSuggestions, cloudId, cloudId))
		}
	}
	return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.CloudProviderNotAvailableErrorMsg, cloudId),
		errors.CloudProviderNotAvailableSuggestions)
}

func (c *clusterCommand) validateEncryptionKey(cmd *cobra.Command, prompt form.Prompt, input validateEncryptionKeyInput) error {
	switch input.Cloud {
	case "aws":
		return c.validateAWSEncryptionKey(cmd, prompt, input)
	case "gcp":
		return c.validateGCPEncryptionKey(cmd, prompt, input)
	default:
		return errors.New(errors.BYOKSupportErrorMsg)
	}
}

func (c *clusterCommand) validateGCPEncryptionKey(cmd *cobra.Command, prompt form.Prompt, input validateEncryptionKeyInput) error {
	ctx := context.Background()
	// The call is idempotent so repeated create commands return the same ID for the same account.
	externalID, err := c.Client.ExternalIdentity.CreateExternalIdentity(ctx, input.Cloud, input.AccountID)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	err = permitBYOKGCP.Execute(buf, struct {
		ExternalIdentity string
	}{
		ExternalIdentity: externalID,
	})
	if err != nil {
		return err
	}
	buf.WriteString("\n\n")
	utils.Println(cmd, buf.String())

	promptMsg := "Please confirm you've authorized the key for this identity: " + externalID
	f := form.New(
		form.Field{ID: "authorized",
			Prompt:    promptMsg,
			IsYesOrNo: true})
	for {
		if err := f.Prompt(cmd, prompt); err != nil {
			utils.ErrPrintln(cmd, errors.FailedToReadConfirmationErrorMsg)
			continue
		}
		if !f.Responses["authorized"].(bool) {
			return errors.Errorf(errors.AuthorizeIdentityErrorMsg, externalID)

		}
		return nil
	}
}

func (c *clusterCommand) validateAWSEncryptionKey(cmd *cobra.Command, prompt form.Prompt, input validateEncryptionKeyInput) error {
	accounts := getEnvironmentsForCloud(input.Cloud, input.MetadataClouds)

	buf := new(bytes.Buffer)
	buf.WriteString(errors.CopyBYOKAWSPermissionsHeaderMsg)
	buf.WriteString("\n\n")
	if err := encryptionKeyPolicy.Execute(buf, accounts); err != nil {
		return errors.New(errors.FailedToRenderKeyPolicyErrorMsg)
	}
	buf.WriteString("\n\n")
	utils.Println(cmd, buf.String())

	promptMsg := "Please confirm you've authorized the key for these accounts: " + strings.Join(accounts, ", ")
	if len(accounts) == 1 {
		promptMsg = "Please confirm you've authorized the key for this account: " + accounts[0]
	}

	f := form.New(form.Field{ID: "authorized", Prompt: promptMsg, IsYesOrNo: true})
	for {
		if err := f.Prompt(cmd, prompt); err != nil {
			utils.ErrPrintln(cmd, errors.FailedToReadConfirmationErrorMsg)
			continue
		}
		if !f.Responses["authorized"].(bool) {
			return errors.Errorf(errors.AuthorizeAccountsErrorMsg, strings.Join(accounts, ", "))
		}
		return nil
	}
}

func getEnvironmentsForCloud(cloudId string, clouds []*ccloudv1.CloudMetadata) []string {
	var environments []string
	for _, cloud := range clouds {
		if cloudId == cloud.Id {
			for _, environment := range cloud.Accounts {
				environments = append(environments, environment.Id)
			}
			break
		}
	}
	return environments
}

func stringToAvailability(s string) (string, error) {
	if modelAvailability, ok := availabilitiesToModel[s]; ok {
		return modelAvailability, nil
	}
	return "", errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidAvailableFlagErrorMsg, s),
		fmt.Sprintf(errors.InvalidAvailableFlagSuggestions, singleZone, multiZone))
}

func stringToSku(skuType string) (ccstructs.Sku, error) {
	sku := ccstructs.Sku(ccstructs.Sku_value[strings.ToUpper(skuType)])
	switch sku {
	case ccstructs.Sku_BASIC, ccstructs.Sku_STANDARD, ccstructs.Sku_DEDICATED:
		break
	default:
		return ccstructs.Sku_UNKNOWN, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidTypeFlagErrorMsg, skuType),
			fmt.Sprintf(errors.InvalidTypeFlagSuggestions, skuBasic, skuStandard, skuDedicated))
	}
	return sku, nil
}

func setCmkClusterConfig(typeString string, cku int32, encryptionKeyID string) *cmkv2.CmkV2ClusterSpecConfigOneOf {
	switch typeString {
	case skuBasic:
		return &cmkv2.CmkV2ClusterSpecConfigOneOf{
			CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
		}
	case skuStandard:
		return &cmkv2.CmkV2ClusterSpecConfigOneOf{
			CmkV2Standard: &cmkv2.CmkV2Standard{Kind: "Standard"},
		}
	case skuDedicated:
		return &cmkv2.CmkV2ClusterSpecConfigOneOf{
			CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: cku, EncryptionKey: &encryptionKeyID},
		}
	default:
		return &cmkv2.CmkV2ClusterSpecConfigOneOf{
			CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
		}
	}
}

func setClusterConfigCku(cluster *cmkv2.CmkV2Cluster, cku int32) {
	cluster.Spec.Config.CmkV2Dedicated.Cku = cku
}

func getKafkaProvisionEstimate(sku ccstructs.Sku) string {
	fmtEstimate := "It may take up to %s for the Kafka cluster to be ready."

	switch sku {
	case ccstructs.Sku_DEDICATED:
		return fmt.Sprintf(fmtEstimate, "1 hour") + " The organization admin will receive an email once the dedicated cluster is provisioned."
	default:
		return fmt.Sprintf(fmtEstimate, "5 minutes")
	}
}
