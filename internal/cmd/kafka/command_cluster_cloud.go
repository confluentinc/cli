package kafka

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/confirm"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"}
	listHumanLabels      = []string{"Id", "Name", "Type", "Provider", "Region", "Availability", "Status"}
	listStructuredLabels = []string{"id", "name", "type", "provider", "region", "availability", "status"}
	basicDescribeFields  = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "ApiEndpoint"}
	describeHumanRenames = map[string]string{
		"NetworkIngress":  "Ingress",
		"NetworkEgress":   "Egress",
		"ServiceProvider": "Provider",
		"EncryptionKeyId": "Encryption Key ID"}
	describeStructuredRenames = map[string]string{
		"Id":                 "id",
		"Name":               "name",
		"Type":               "type",
		"ClusterSize":        "cluster_size",
		"PendingClusterSize": "pending_cluster_size",
		"NetworkIngress":     "ingress",
		"NetworkEgress":      "egress",
		"Storage":            "storage",
		"ServiceProvider":    "provider",
		"Region":             "region",
		"Availability":       "availability",
		"Status":             "status",
		"Endpoint":           "endpoint",
		"ApiEndpoint":        "api_endpoint",
		"EncryptionKeyId":    "encryption_key_id"}
)

const (
	singleZone   = "single-zone"
	multiZone    = "multi-zone"
	skuBasic     = "basic"
	skuStandard  = "standard"
	skuDedicated = "dedicated"
)

type clusterCommand struct {
	*pcmd.AuthenticatedCLICommand
	prerunner pcmd.PreRunner
}

type describeStruct struct {
	Id                 string
	Name               string
	Type               string
	ClusterSize        int32
	PendingClusterSize int32
	NetworkIngress     int32
	NetworkEgress      int32
	Storage            int32
	ServiceProvider    string
	Region             string
	Availability       string
	Status             string
	Endpoint           string
	ApiEndpoint        string
	EncryptionKeyId    string
}

// NewClusterCommand returns the Cobra command for Kafka cluster.
func NewClusterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "cluster",
			Short: "Manage Kafka clusters.",
		}, prerunner)
	cmd := &clusterCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	cmd.init()
	return cmd.Command
}

func (c *clusterCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka clusters.",
		RunE:  pcmd.NewCLIRunE(c.list),
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Kafka cluster.",
		RunE:  pcmd.NewCLIRunE(c.create),
		Args:  cobra.ExactArgs(1),
	}

	createCmd.Flags().String("cloud", "", "Cloud provider ID (e.g. 'aws' or 'gcp').")
	createCmd.Flags().String("region", "", "Cloud region ID for cluster (e.g. 'us-west-2').")
	check(createCmd.MarkFlagRequired("cloud"))
	check(createCmd.MarkFlagRequired("region"))
	createCmd.Flags().String("availability", singleZone, fmt.Sprintf("Availability of the cluster. Allowed Values: %s, %s.", singleZone, multiZone))
	createCmd.Flags().String("type", skuBasic, fmt.Sprintf("Type of the Kafka cluster. Allowed values: %s, %s, %s.", skuBasic, skuStandard, skuDedicated))
	createCmd.Flags().Int("cku", 0, "Number of Confluent Kafka Units (non-negative). Required for Kafka clusters of type 'dedicated'.")
	createCmd.Flags().String("encryption-key", "", "Encryption Key ID (e.g. for Amazon Web Services, the Amazon Resource Name of the key).")
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Describe a Kafka cluster.",
		RunE:  pcmd.NewCLIRunE(c.describe),
		Args:  cobra.ExactArgs(1),
	}
	describeCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	describeCmd.Flags().SortFlags = false
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Kafka cluster.",
		RunE:  pcmd.NewCLIRunE(c.update),
		Args:  cobra.ExactArgs(1),
	}
	updateCmd.Flags().String("name", "", "Name of the Kafka cluster.")
	updateCmd.Flags().Int("cku", 0, "Number of Confluent Kafka Units (non-negative). For Kafka clusters of type 'dedicated' only.")
	updateCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a Kafka cluster.",
		RunE:  pcmd.NewCLIRunE(c.delete),
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(deleteCmd)
	c.AddCommand(&cobra.Command{
		Use:   "use <id>",
		Short: "Make the Kafka cluster active for use in other commands.",
		RunE:  pcmd.NewCLIRunE(c.use),
		Args:  cobra.ExactArgs(1),
	})
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.Kafka.List(context.Background(), req)
	if err != nil {
		return err
	}
	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, cluster := range clusters {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if cluster.Id == c.Context.KafkaClusterContext.GetActiveKafkaClusterId() {
				cluster.Id = fmt.Sprintf("* %s", cluster.Id)
			} else {
				cluster.Id = fmt.Sprintf("  %s", cluster.Id)
			}
		}
		outputWriter.AddElement(convertClusterToDescribeStruct(cluster))
	}
	return outputWriter.Out()
}

var stdin io.ReadWriter = os.Stdin
var stdout io.ReadWriter = os.Stdout

func (c *clusterCommand) create(cmd *cobra.Command, args []string) error {
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
	err = checkCloudAndRegion(cloud, region, clouds)
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
	typeString, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	sku, err := stringToSku(typeString)
	if err != nil {
		return err
	}
	encryptionKeyID, err := cmd.Flags().GetString("encryption-key")
	if err != nil {
		return err
	}
	if encryptionKeyID != "" {
		if err := c.validateEncryptionKey(cloud, clouds); err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	cfg := &schedv1.KafkaClusterConfig{
		AccountId:       c.EnvironmentId(),
		Name:            args[0],
		ServiceProvider: cloud,
		Region:          region,
		Durability:      availability,
		Deployment:      &schedv1.Deployment{Sku: sku},
		EncryptionKeyId: encryptionKeyID,
	}
	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetInt("cku")
		if err != nil {
			return err
		}
		if sku != productv1.Sku_DEDICATED {
			return errors.New(errors.CKUOnlyForDedicatedErrorMsg)
		}
		if cku <= 0 {
			return errors.New(errors.CKUMoreThanZeroErrorMsg)
		}
		cfg.Cku = int32(cku)
	}

	cluster, err := c.Client.Kafka.Create(context.Background(), cfg)
	if err != nil {
		// TODO: don't swallow validation errors (reportedly separately)
		return err
	}
	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}
	if outputFormat == output.Human.String() {
		pcmd.ErrPrintln(cmd, errors.KafkaClusterTime)
	}
	return outputKafkaClusterDescription(cmd, cluster)
}

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

func (c *clusterCommand) validateEncryptionKey(cloud string, clouds []*schedv1.CloudMetadata) error {
	accounts := getEnvironmentsForCloud(cloud, clouds)
	var buf bytes.Buffer
	buf.WriteString(errors.CopyBYOKPermissionsHeaderMsg)
	buf.WriteString("\n\n")
	if err := encryptionKeyPolicy.Execute(&buf, accounts); err != nil {
		return errors.New(errors.FailedToRenderKeyPolicyErrorMsg)
	}
	buf.WriteString("\n\n")
	ok, err := confirm.Do(
		stdout,
		stdin,
		buf.String(),
	)
	if err != nil {
		return errors.New(errors.FailedToReadConfirmationErrorMsg)
	}
	if !ok {
		return errors.Errorf(errors.AuthorizeAccountsErrorMsg, strings.Join(accounts, ", "))
	}
	return nil
}

func stringToAvailability(s string) (schedv1.Durability, error) {
	if s == singleZone {
		return schedv1.Durability_LOW, nil
	} else if s == multiZone {
		return schedv1.Durability_HIGH, nil
	}
	return schedv1.Durability_LOW, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidAvailableFlagErrorMsg, s),
		fmt.Sprintf(errors.InvalidAvailableFlagSuggestions, singleZone, multiZone))
}

func stringToSku(s string) (productv1.Sku, error) {
	sku := productv1.Sku(productv1.Sku_value[strings.ToUpper(s)])
	switch sku {
	case productv1.Sku_BASIC, productv1.Sku_STANDARD, productv1.Sku_DEDICATED:
		break
	default:
		return productv1.Sku_UNKNOWN, errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidTypeFlagErrorMsg, s),
			fmt.Sprintf(errors.InvalidTypeFlagSuggestions, skuBasic, skuStandard, skuDedicated))
	}
	return sku, nil
}

func (c *clusterCommand) describe(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	cluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}
	return outputKafkaClusterDescription(cmd, cluster)
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("cku") {
		return errors.New(errors.NameOrCKUFlagErrorMsg)
	}
	req := &schedv1.KafkaCluster{
		AccountId: c.EnvironmentId(),
		Id:        args[0],
	}
	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		if name == "" {
			return errors.New(errors.NonEmptyNameErrorMsg)
		}
		req.Name = name
	} else {
		currentCluster, err := c.Client.Kafka.Describe(context.Background(), req)
		if err != nil {
			return err
		}
		req.Name = currentCluster.Name
	}
	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetInt("cku")
		if err != nil {
			return err
		}
		if cku <= 0 {
			return errors.New(errors.CKUMoreThanZeroErrorMsg)
		}
		req.Cku = int32(cku)
	}
	updatedCluster, err := c.Client.Kafka.Update(context.Background(), req)
	if err != nil {
		return err
	}
	return outputKafkaClusterDescription(cmd, updatedCluster)
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	err := c.Client.Kafka.Delete(context.Background(), req)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}
	err = c.Context.RemoveKafkaClusterConfig(args[0])
	if err != nil {
		return err
	}
	pcmd.Printf(cmd, errors.KafkaClusterDeletedMsg, args[0])
	return nil
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterID := args[0]

	_, err := c.Context.FindKafkaCluster(cmd, clusterID)
	if err != nil {
		err = errors.CatchKafkaNotFoundError(err, clusterID)
		return err
	}
	err = c.Context.SetActiveKafkaCluster(cmd, clusterID)
	if err != nil {
		return err
	}
	cmd.PrintErrf(errors.UseKafkaClusterMsg, clusterID, c.Context.GetCurrentEnvironmentId())
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func checkCloudAndRegion(cloudId string, regionId string, clouds []*schedv1.CloudMetadata) error {
	for _, cloud := range clouds {
		if cloudId == cloud.Id {
			for _, region := range cloud.Regions {
				if regionId == region.Id {
					if region.IsSchedulable {
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

func getEnvironmentsForCloud(cloudId string, clouds []*schedv1.CloudMetadata) []string {
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

func outputKafkaClusterDescription(cmd *cobra.Command, cluster *schedv1.KafkaCluster) error {
	return output.DescribeObject(cmd, convertClusterToDescribeStruct(cluster), getKafkaClusterDescribeFields(cluster), describeHumanRenames, describeStructuredRenames)
}

func convertClusterToDescribeStruct(cluster *schedv1.KafkaCluster) *describeStruct {
	return &describeStruct{
		Id:                 cluster.Id,
		Name:               cluster.Name,
		Type:               cluster.Deployment.Sku.String(),
		ClusterSize:        cluster.Cku,
		PendingClusterSize: cluster.PendingCku,
		NetworkIngress:     cluster.NetworkIngress,
		NetworkEgress:      cluster.NetworkEgress,
		Storage:            cluster.Storage,
		ServiceProvider:    cluster.ServiceProvider,
		Region:             cluster.Region,
		Availability:       cluster.Durability.String(),
		Status:             cluster.Status.String(),
		Endpoint:           cluster.Endpoint,
		ApiEndpoint:        cluster.ApiEndpoint,
		EncryptionKeyId:    cluster.EncryptionKeyId,
	}
}

func getKafkaClusterDescribeFields(cluster *schedv1.KafkaCluster) []string {
	describeFields := basicDescribeFields
	if isDedicated(cluster) {
		describeFields = append(describeFields, "ClusterSize")
		if isExpanding(cluster) {
			describeFields = append(describeFields, "PendingClusterSize")
		}
		if cluster.EncryptionKeyId != "" {
			describeFields = append(describeFields, "EncryptionKeyId")
		}
	}
	return describeFields
}

func isDedicated(cluster *schedv1.KafkaCluster) bool {
	return cluster.Deployment.Sku == productv1.Sku_DEDICATED
}

func isExpanding(cluster *schedv1.KafkaCluster) bool {
	return cluster.Status == schedv1.ClusterStatus_EXPANDING || cluster.PendingCku > cluster.Cku
}
