package kafka

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	pconv "github.com/confluentinc/cli/internal/pkg/name-conversions"
	"github.com/confluentinc/cli/internal/pkg/output"
)

const (
	skuBasic     = "basic"
	skuStandard  = "standard"
	skuDedicated = "dedicated"
)

var permitBYOKGCP = template.Must(template.New("byok_gcp_permissions").Parse(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  {{.ExternalIdentity}}`))

func (c *clusterCommand) newCreateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "create <name>",
		Short:       "Create a Kafka cluster.",
		Long:        "Create a Kafka cluster.\n\nNote: You cannot use this command to create a cluster that is configured with AWS PrivateLink. You must use the UI to create a cluster of that configuration.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.create,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new dedicated cluster that uses a customer-managed encryption key in GCP:",
				Code: `confluent kafka cluster create sales092020 --cloud gcp --region asia-southeast1 --type dedicated --cku 1 --encryption-key "projects/PROJECT_NAME/locations/LOCATION/keyRings/KEY_RING/cryptoKeys/KEY_NAME"`,
			},
			examples.Example{
				Text: "Create a new dedicated cluster that uses a customer-managed encryption key in AWS:",
				Code: "confluent kafka cluster create my-cluster --cloud aws --region us-west-2 --type dedicated --cku 1 --byok cck-a123z",
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
	cmd.Flags().String("encryption-key", "", "Resource ID of the Cloud Key Management Service key (GCP only).")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddByokKeyFlag(cmd, c.AuthenticatedCLICommand)
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	clouds, err := c.Client.EnvironmentMetadata.Get()
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

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	var encryptionKey string
	if cmd.Flags().Changed("encryption-key") {
		if cloud != "gcp" {
			return errors.New(errors.EncryptionKeySupportErrorMsg)
		}

		encryptionKey, err = cmd.Flags().GetString("encryption-key")
		if err != nil {
			return err
		}

		if err := c.validateGcpEncryptionKey(cloud, environmentId); err != nil {
			environmentId, err = pconv.ConvertEnvironmentNameToId(environmentId, c.V2Client)
			if err != nil {
				return err
			}
			if err = c.validateGcpEncryptionKey(cloud, environmentId); err != nil {
				return err
			}
		}
	}

	byok, err := cmd.Flags().GetString("byok")
	if err != nil {
		return err
	}

	var keyGlobalObjectReference *cmkv2.GlobalObjectReference
	if byok != "" {
		key, httpResp, err := c.V2Client.GetByokKey(byok)
		if err != nil {
			return errors.CatchByokKeyNotFoundError(err, httpResp)
		}
		keyGlobalObjectReference = &cmkv2.GlobalObjectReference{Id: key.GetId()}
	}

	createCluster := cmkv2.CmkV2Cluster{
		Spec: &cmkv2.CmkV2ClusterSpec{
			Environment:  &cmkv2.EnvScopedObjectReference{Id: environmentId},
			DisplayName:  cmkv2.PtrString(args[0]),
			Cloud:        cmkv2.PtrString(cloud),
			Region:       cmkv2.PtrString(region),
			Availability: cmkv2.PtrString(availability),
			Config:       setCmkClusterConfig(clusterType, 1, encryptionKey),
			Byok:         keyGlobalObjectReference,
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
		environmentId, err = pconv.ConvertEnvironmentNameToId(environmentId, c.V2Client)
		if err != nil {
			return err
		}
		kafkaCluster, httpResp, err = c.V2Client.CreateKafkaCluster(createCluster)
		return errors.CatchClusterConfigurationNotValidError(err, httpResp)
	}

	if output.GetFormat(cmd) == output.Human {
		output.ErrPrintln(getKafkaProvisionEstimate(sku))
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

func (c *clusterCommand) validateGcpEncryptionKey(cloud, accountId string) error {
	// The call is idempotent so repeated create commands return the same ID for the same account.
	externalID, err := c.Client.ExternalIdentity.CreateExternalIdentity(cloud, accountId)
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
	output.Println(buf.String())

	promptMsg := "Please confirm you've authorized the key for this identity: " + externalID
	f := form.New(
		form.Field{
			ID:        "authorized",
			Prompt:    promptMsg,
			IsYesOrNo: true,
		})
	for {
		if err := f.Prompt(form.NewPrompt()); err != nil {
			output.ErrPrintln(errors.FailedToReadConfirmationErrorMsg)
			continue
		}
		if !f.Responses["authorized"].(bool) {
			return errors.Errorf(errors.AuthorizeIdentityErrorMsg, externalID)
		}
		return nil
	}
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
		var encryptionPtr *string
		if encryptionKeyID != "" {
			encryptionPtr = &encryptionKeyID
		}

		return &cmkv2.CmkV2ClusterSpecConfigOneOf{
			CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: cku, EncryptionKey: encryptionPtr},
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
