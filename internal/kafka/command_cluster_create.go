package kafka

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/spf13/cobra"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/v3/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/kafka"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

const (
	skuBasic      = "basic"
	skuStandard   = "standard"
	skuEnterprise = "enterprise"
	skuDedicated  = "dedicated"
)

var permitBYOKGCP = template.Must(template.New("byok_gcp_permissions").Parse(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  {{.ExternalIdentity}}`))

func (c *clusterCommand) newCreateCommand() *cobra.Command {
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
	pcmd.AddByokKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
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
			return fmt.Errorf("BYOK via `--encryption-key` is only available for GCP. Use `confluent byok create` to register AWS and Azure keys.")
		}

		encryptionKey, err = cmd.Flags().GetString("encryption-key")
		if err != nil {
			return err
		}

		if err := c.validateGcpEncryptionKey(cloud, environmentId); err != nil {
			return err
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

	createCluster := cmkv2.CmkV2Cluster{Spec: &cmkv2.CmkV2ClusterSpec{
		Environment:  &cmkv2.EnvScopedObjectReference{Id: environmentId},
		DisplayName:  cmkv2.PtrString(args[0]),
		Cloud:        cmkv2.PtrString(cloud),
		Region:       cmkv2.PtrString(region),
		Availability: cmkv2.PtrString(availability),
		Config:       setCmkClusterConfig(clusterType, 1, encryptionKey),
		Byok:         keyGlobalObjectReference,
	}}

	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetInt("cku")
		if err != nil {
			return err
		}
		if clusterType != skuDedicated {
			return errors.NewErrorWithSuggestions("the `--cku` flag can only be used when creating a dedicated Kafka cluster", "Specify a dedicated cluster with `--type`.")
		}
		if cku <= 0 {
			return fmt.Errorf(errors.CkuMoreThanZeroErrorMsg)
		}
		setClusterConfigCku(&createCluster, int32(cku))
	}

	kafkaCluster, httpResp, err := c.V2Client.CreateKafkaCluster(createCluster)
	if err != nil {
		return catchClusterConfigurationNotValidError(err, httpResp, cloud, region)
	}

	if output.GetFormat(cmd) == output.Human {
		output.ErrPrintln(c.Config.EnableColor, getKafkaProvisionEstimate(sku))
	}

	return c.outputKafkaClusterDescription(cmd, &kafkaCluster, false)
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
	output.Println(c.Config.EnableColor, buf.String())

	promptMsg := "Please confirm you've authorized the key for this identity: " + externalID
	f := form.New(
		form.Field{
			ID:        "authorized",
			Prompt:    promptMsg,
			IsYesOrNo: true,
		})
	for {
		if err := f.Prompt(form.NewPrompt()); err != nil {
			output.ErrPrintln(c.Config.EnableColor, "BYOK error: failed to read your confirmation")
			continue
		}
		if !f.Responses["authorized"].(bool) {
			return fmt.Errorf("BYOK error: please authorize the key for the identity (%s)", externalID)
		}
		return nil
	}
}

func stringToAvailability(s string) (string, error) {
	if modelAvailability, ok := availabilitiesToModel[s]; ok {
		return modelAvailability, nil
	}
	return "", errors.NewErrorWithSuggestions(
		fmt.Sprintf("invalid value \"%s\" for `--availability` flag", s),
		fmt.Sprintf("Allowed values for `--availability` flag are: %s, %s.", singleZone, multiZone),
	)
}

func stringToSku(skuType string) (ccstructs.Sku, error) {
	sku := ccstructs.Sku(ccstructs.Sku_value[strings.ToUpper(skuType)])
	switch sku {
	case ccstructs.Sku_BASIC, ccstructs.Sku_STANDARD, ccstructs.Sku_ENTERPRISE, ccstructs.Sku_DEDICATED:
		break
	default:
		return ccstructs.Sku_UNKNOWN, errors.NewErrorWithSuggestions(
			fmt.Sprintf("invalid value \"%s\" for `--type` flag", skuType),
			fmt.Sprintf("Allowed values for `--type` flag are: %s.", utils.ArrayToCommaDelimitedString(kafka.Types, "and")),
		)
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
	case skuEnterprise:
		return &cmkv2.CmkV2ClusterSpecConfigOneOf{CmkV2Enterprise: &cmkv2.CmkV2Enterprise{Kind: "Enterprise"}}
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

func catchClusterConfigurationNotValidError(err error, r *http.Response, cloud, region string) error {
	if err == nil || r == nil {
		return err
	}

	err = errors.CatchCCloudV2Error(err, r)

	if err.Error() == "Service provider must be set to AWS, GCP or AZURE." {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`"%s" is not an available cloud provider`, cloud),
			"To view a list of available cloud providers and regions, use `confluent kafka region list`.",
		)
	}
	if err.Error() == "Unable to schedule given the cloud and/or region in request is invalid or unavailable" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`"%s" is not an available region for "%s"`, region, cloud),
			fmt.Sprintf("To view a list of available regions for \"%s\", use `confluent kafka region list --cloud %s`.", cloud, cloud),
		)
	}
	if strings.Contains(err.Error(), "CKU must be greater") {
		return fmt.Errorf("CKU must be greater than 1 for multi-zone dedicated clusters")
	}
	if strings.Contains(err.Error(), "Durability must be HIGH for an Enterprise cluster") {
		return fmt.Errorf(`availability must be "multi-zone" for enterprise clusters`)
	}

	return err
}
