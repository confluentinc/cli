package kafka

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"

	"github.com/c-bata/go-prompt"
	"github.com/spf13/cobra"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	pkafka "github.com/confluentinc/cli/internal/pkg/kafka"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var (
	listFields                         = []string{"Id", "Name", "Type", "ServiceProvider", "Region", "Availability", "Status"}
	listHumanLabels                    = []string{"Id", "Name", "Type", "Provider", "Region", "Availability", "Status"}
	listStructuredLabels               = []string{"id", "name", "type", "provider", "region", "availability", "status"}
	basicDescribeFields                = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "RestEndpoint"}
	basicDescribeFieldsWithApiEndpoint = []string{"Id", "Name", "Type", "NetworkIngress", "NetworkEgress", "Storage", "ServiceProvider", "Availability", "Region", "Status", "Endpoint", "ApiEndpoint", "RestEndpoint"}
	describeHumanRenames               = map[string]string{
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
		"EncryptionKeyId":    "encryption_key_id",
		"RestEndpoint":       "rest_endpoint",
	}
	durabilityToAvaiablityNameMap = map[string]string{
		"LOW":  singleZone,
		"HIGH": multiZone,
	}
)

const (
	singleZone   = "single-zone"
	multiZone    = "multi-zone"
	skuBasic     = "basic"
	skuStandard  = "standard"
	skuDedicated = "dedicated"
)

type clusterCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	logger              *log.Logger
	prerunner           pcmd.PreRunner
	completableChildren []*cobra.Command
	analyticsClient     analytics.Client
}

type describeStruct struct {
	Id                 string
	Name               string
	Type               string
	ClusterSize        int32
	PendingClusterSize int32
	NetworkIngress     int32
	NetworkEgress      int32
	Storage            string
	ServiceProvider    string
	Region             string
	Availability       string
	Status             string
	Endpoint           string
	ApiEndpoint        string
	EncryptionKeyId    string
	RestEndpoint       string
}

type describeStructWithKAPI struct {
	Id                 string
	Name               string
	Type               string
	ClusterSize        int32
	PendingClusterSize int32
	NetworkIngress     int32
	NetworkEgress      int32
	Storage            string
	ServiceProvider    string
	Region             string
	Availability       string
	Status             string
	Endpoint           string
	EncryptionKeyId    string
	RestEndpoint       string
	KAPI               string
}

// NewClusterCommand returns the command for Kafka cluster.
func NewClusterCommand(cfg *v1.Config, prerunner pcmd.PreRunner, analyticsClient analytics.Client) *clusterCommand {
	cmd := &cobra.Command{
		Use:         "cluster",
		Short:       "Manage Kafka clusters.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &clusterCommand{
		prerunner:       prerunner,
		analyticsClient: analyticsClient,
	}
	if cfg.IsCloudLogin() {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner, ClusterSubcommandFlags)
	} else {
		c.AuthenticatedStateFlagCommand = pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, OnPremClusterSubcommandFlags)
	}

	c.init(cfg)

	return c
}

func (c *clusterCommand) init(cfg *v1.Config) {
	listCmd := &cobra.Command{
		Use:  "list",
		Args: cobra.NoArgs,
	}
	if cfg.IsCloudLogin() {
		listCmd.Short = "List Kafka clusters."
		listCmd.RunE = pcmd.NewCLIRunE(c.list)
		listCmd.Flags().Bool("all", false, "List clusters across all environments.")
	} else {
		listCmd.Short = "List registered Kafka clusters."
		listCmd.Long = "List Kafka clusters that are registered with the MDS cluster registry."
		listCmd.RunE = pcmd.NewCLIRunE(c.onPremList)
	}
	output.AddFlag(listCmd)
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Kafka cluster.",
		Long:  "Create a Kafka cluster.\n\nNote: You cannot use this command to create a cluster that is configured with AWS PrivateLink. You must use the UI to create a cluster of that configuration.",
		Args:  cobra.ExactArgs(1),
		RunE: pcmd.NewCLIRunE(func(cmd *cobra.Command, args []string) error {
			return c.create(cmd, args, form.NewPrompt(os.Stdin))
		}),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a new dedicated cluster that uses a customer-managed encryption key in AWS:",
				Code: `confluent kafka cluster create sales092020 --cloud "aws" --region "us-west-2" --type "dedicated" --cku 1 --encryption-key "arn:aws:kms:us-west-2:111122223333:key/1234abcd-12ab-34cd-56ef-1234567890ab"`,
			},
			examples.Example{
				Text: "For more information, see https://docs.confluent.io/current/cloud/clusters/byok-encrypted-clusters.html.",
			},
		),
	}

	createCmd.Flags().String("cloud", "", "Cloud provider ID (e.g. 'aws' or 'gcp').")
	createCmd.Flags().String("region", "", "Cloud region ID for cluster (e.g. 'us-west-2').")
	check(createCmd.MarkFlagRequired("cloud"))
	check(createCmd.MarkFlagRequired("region"))
	createCmd.Flags().String("availability", singleZone, fmt.Sprintf("Availability of the cluster. Allowed Values: %s, %s.", singleZone, multiZone))
	createCmd.Flags().String("type", skuBasic, fmt.Sprintf("Type of the Kafka cluster. Allowed values: %s, %s, %s.", skuBasic, skuStandard, skuDedicated))
	createCmd.Flags().Int("cku", 0, "Number of Confluent Kafka Units (non-negative). Required for Kafka clusters of type 'dedicated'.")
	createCmd.Flags().String("encryption-key", "", "Encryption Key ID (e.g. for Amazon Web Services, the Amazon Resource Name of the key).")
	output.AddFlag(createCmd)
	c.AddCommand(createCmd)

	describeCmd := &cobra.Command{
		Use:         "describe [id]",
		Short:       "Describe a Kafka cluster.",
		Long:        "Describe the Kafka cluster specified with the ID argument, or describe the active cluster for the current context.",
		Args:        cobra.MaximumNArgs(1),
		RunE:        pcmd.NewCLIRunE(c.describe),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	output.AddFlag(describeCmd)
	describeCmd.Flags().Bool("all", false, "List all properties of a Kafka cluster.")
	c.AddCommand(describeCmd)

	updateCmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update a Kafka cluster.",
		Args:  cobra.ExactArgs(1),
		RunE: pcmd.NewCLIRunE(func(cmd *cobra.Command, args []string) error {
			return c.update(cmd, args, form.NewPrompt(os.Stdin))
		}),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Change a cluster's name and expand its CKU count:",
				Code: `confluent kafka cluster update lkc-abc123 --name "Cool Cluster" --cku 3`,
			},
		),
	}
	updateCmd.Flags().String("name", "", "Name of the Kafka cluster.")
	updateCmd.Flags().Int("cku", 0, "Number of Confluent Kafka Units (non-negative). For Kafka clusters of type 'dedicated' only. When shrinking a cluster, you can reduce capacity one CKU at a time.")
	output.AddFlag(updateCmd)
	c.AddCommand(updateCmd)

	deleteCmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a Kafka cluster.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.delete),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	c.AddCommand(deleteCmd)
	useCmd := &cobra.Command{
		Use:         "use <id>",
		Short:       "Make the Kafka cluster active for use in other commands.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.use),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
	c.AddCommand(useCmd)
	c.completableChildren = []*cobra.Command{deleteCmd, describeCmd, updateCmd, useCmd}
}

func (c *clusterCommand) list(cmd *cobra.Command, _ []string) error {
	listAllClusters, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}
	var clusters []*schedv1.KafkaCluster
	if listAllClusters {
		environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
		if err != nil {
			return err
		}

		for _, env := range environments {
			clustersOfEnv, err := pkafka.ListKafkaClusters(c.Client, env.Id)
			if err != nil {
				return err
			}

			clusters = append(clusters, clustersOfEnv...)
		}
	} else {
		clusters, err = pkafka.ListKafkaClusters(c.Client, c.EnvironmentId())
		if err != nil {
			return err
		}
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
		if err := c.validateEncryptionKey(cmd, prompt, validateEncryptionKeyInput{
			Cloud:          cloud,
			MetadataClouds: clouds,
			AccountID:      c.EnvironmentId(),
		}); err != nil {
			return err
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
		utils.ErrPrintln(cmd, errors.KafkaClusterTime)
	}
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, cluster.Id)
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

type validateEncryptionKeyInput struct {
	Cloud          string
	MetadataClouds []*schedv1.CloudMetadata
	AccountID      string
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

var permitBYOKGCP = template.Must(template.New("byok_gcp_permissions").Parse(`Create a role with these permissions, add the identity as a member of your key, and grant your role to the member:

Permissions:
  - cloudkms.cryptoKeyVersions.useToDecrypt
  - cloudkms.cryptoKeyVersions.useToEncrypt
  - cloudkms.cryptoKeys.get

Identity:
  {{.ExternalIdentity}}`))

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
	all, err := cmd.Flags().GetBool("all")
	if err != nil {
		return err
	}
	lkc, err := c.getLkcForDescribe(args)
	if err != nil {
		return err
	}
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: lkc}
	cluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}
	return outputKafkaClusterDescriptionWithKAPI(cmd, cluster, all)
}

func (c *clusterCommand) getLkcForDescribe(args []string) (string, error) {
	if len(args) > 0 {
		return args[0], nil
	}
	lkc := c.Config.Context().KafkaClusterContext.GetActiveKafkaClusterId()
	if lkc == "" {
		return "", errors.NewErrorWithSuggestions(errors.NoKafkaSelectedErrorMsg, errors.NoKafkaForDescribeSuggestions)
	}
	return lkc, nil
}

func (c *clusterCommand) update(cmd *cobra.Command, args []string, prompt form.Prompt) error {
	if !cmd.Flags().Changed("name") && !cmd.Flags().Changed("cku") {
		return errors.New(errors.NameOrCKUFlagErrorMsg)
	}
	clusterID := args[0]
	req := &schedv1.KafkaCluster{
		AccountId: c.EnvironmentId(),
		Id:        clusterID,
	}
	currentCluster, err := c.Client.Kafka.Describe(context.Background(), req)
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterID), errors.ChooseRightEnvironmentSuggestions)
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
		req.Name = currentCluster.Name
	}
	req.Cku, err = c.validateResize(cmd, currentCluster, prompt)
	if err != nil {
		return err
	}

	updatedCluster, err := c.Client.Kafka.Update(context.Background(), req)
	if err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), errors.KafkaClusterUpdateFailedSuggestions)
	}
	return outputKafkaClusterDescription(cmd, updatedCluster)
}

func (c *clusterCommand) validateResize(cmd *cobra.Command, currentCluster *schedv1.KafkaCluster, prompt form.Prompt) (int32, error) {
	if cmd.Flags().Changed("cku") {
		cku, err := cmd.Flags().GetInt("cku")
		if err != nil {
			return currentCluster.Cku, err
		}
		// Ensure the cluster is a Dedicated Cluster
		if !isDedicated(currentCluster) {
			return currentCluster.Cku, errors.Errorf("error updating kafka cluster: %v", errors.ClusterResizeNotSupported)
		}
		// Durability Checks
		if currentCluster.Durability == schedv1.Durability_HIGH && cku <= 1 {
			return currentCluster.Cku, errors.New(errors.CKUMoreThanOneErrorMsg)
		}
		if cku <= 0 {
			return currentCluster.Cku, errors.New(errors.CKUMoreThanZeroErrorMsg)
		}
		// Cluster can't be resized while it's provisioning or being expanded already.
		// Name _can_ be changed during these times, though.
		err = isClusterResizeInProgress(currentCluster)
		if err != nil {
			return currentCluster.Cku, err
		}
		//If shrink
		if int32(cku) < currentCluster.Cku {
			// metrics api auth via jwt
			shouldPrompt, errFromSmallWindowMetrics := c.validateKafkaClusterMetrics(context.Background(), int32(cku), currentCluster, true)
			if errFromSmallWindowMetrics != nil && !shouldPrompt {
				return currentCluster.Cku, fmt.Errorf("cluster shrink validation error: \n%v", errFromSmallWindowMetrics)
			}
			promptMessage := ""
			if shouldPrompt {
				promptMessage = fmt.Sprintf("\n%v\n", errFromSmallWindowMetrics)
			}
			_, errFromLargeWindowMetrics := c.validateKafkaClusterMetrics(context.Background(), int32(cku), currentCluster, false)
			if errFromLargeWindowMetrics != nil {
				promptMessage += fmt.Sprintf("\n%v\n", errFromLargeWindowMetrics)
			}
			if promptMessage != "" {
				ok, err := confirmShrink(cmd, prompt, promptMessage)
				if !ok || err != nil {
					return currentCluster.Cku, err
				} else {
					return int32(cku), nil
				}
			}
		}
		return int32(cku), nil
	}
	return currentCluster.Cku, nil
}

func (c *clusterCommand) validateKafkaClusterMetrics(ctx context.Context, cku int32, currentCluster *schedv1.KafkaCluster, isLatestMetric bool) (bool, error) {
	var window string
	if isLatestMetric {
		window = "15 min"
	} else {
		window = "3 days"
	}
	requiredPartitionCount, requiredStorageLimit, err := c.getUsageLimit(ctx, uint32(cku))
	if err != nil {
		c.logger.Warn("Could not retrieve usage limits ", err)
		return false, errors.New("Could not retrieve usage limits to validate request to shrink cluster.")
	}
	errorMessage := errors.Errorf("Looking at metrics in the last %s window:", window)
	shouldPrompt := true
	isValidPartitionCountErr := c.validatePartitionCount(currentCluster.Id, requiredPartitionCount, isLatestMetric, cku)
	if isValidPartitionCountErr != nil {
		errorMessage = errors.Errorf("%v \n %v", errorMessage.Error(), isValidPartitionCountErr.Error())
		shouldPrompt = false
	}
	var isValidStorageLimitErr error
	if !currentCluster.InfiniteStorage {
		isValidStorageLimitErr = c.validateStorageLimit(currentCluster.Id, requiredStorageLimit, isLatestMetric, cku)
		if isValidStorageLimitErr != nil {
			errorMessage = errors.Errorf("%v \n %v", errorMessage.Error(), isValidStorageLimitErr.Error())
			shouldPrompt = false
		}
	}
	// Get Cluster Load Metric
	isValidLoadErr := c.validateClusterLoad(currentCluster.Id, isLatestMetric)
	if isValidLoadErr != nil {
		errorMessage = errors.Errorf("%v \n %v", errorMessage.Error(), isValidLoadErr)
	}
	if isValidStorageLimitErr == nil && isValidLoadErr == nil && isValidPartitionCountErr == nil {
		return false, nil
	}
	return shouldPrompt, errorMessage
}

func confirmShrink(cmd *cobra.Command, prompt form.Prompt, promptMessage string) (bool, error) {
	f := form.New(form.Field{ID: "proceed", Prompt: fmt.Sprintf("Validated cluster metrics and found that: %s\nDo you want to proceed with shrinking your kafka cluster?", promptMessage), IsYesOrNo: true})
	if err := f.Prompt(cmd, prompt); err != nil {
		return false, errors.New(errors.FailedToReadClusterResizeConfirmationErrorMsg)
	}
	if !f.Responses["proceed"].(bool) {
		utils.Println(cmd, "Not proceeding with kafka cluster shrink")
		return false, nil
	}
	return true, nil
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
	utils.Printf(cmd, errors.KafkaClusterDeletedMsg, args[0])
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, args[0])
	return nil
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterID := args[0]

	if _, err := c.Context.FindKafkaCluster(clusterID); err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterID), errors.ChooseRightEnvironmentSuggestions)
	}

	if err := c.Context.SetActiveKafkaCluster(clusterID); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.UseKafkaClusterMsg, clusterID, c.Context.GetCurrentEnvironmentId())
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
	return output.DescribeObject(cmd, convertClusterToDescribeStruct(cluster), getKafkaClusterDescribeFields(cluster, basicDescribeFieldsWithApiEndpoint), describeHumanRenames, describeStructuredRenames)
}

func outputKafkaClusterDescriptionWithKAPI(cmd *cobra.Command, cluster *schedv1.KafkaCluster, all bool) error {
	fields := basicDescribeFields
	structureRenames := describeStructuredRenames
	delete(structureRenames, "ApiEndpoint")
	if all { // expose KAPI when --all flag is set
		fields = append(fields, "KAPI")
		structureRenames["KAPI"] = "kapi"
	}
	return output.DescribeObject(cmd, convertClusterToDescribeStructWithKAPI(cluster), getKafkaClusterDescribeFields(cluster, fields), describeHumanRenames, structureRenames)
}

func convertClusterToDescribeStruct(cluster *schedv1.KafkaCluster) *describeStruct {
	clusterStorage := strconv.Itoa(int(cluster.Storage))
	if clusterStorage == "-1" || cluster.InfiniteStorage {
		clusterStorage = "Infinite"
	}

	return &describeStruct{
		Id:                 cluster.Id,
		Name:               cluster.Name,
		Type:               cluster.Deployment.Sku.String(),
		ClusterSize:        cluster.Cku,
		PendingClusterSize: cluster.PendingCku,
		NetworkIngress:     cluster.NetworkIngress,
		NetworkEgress:      cluster.NetworkEgress,
		Storage:            clusterStorage,
		ServiceProvider:    cluster.ServiceProvider,
		Region:             cluster.Region,
		Availability:       durabilityToAvaiablityNameMap[cluster.Durability.String()],
		Status:             cluster.Status.String(),
		Endpoint:           cluster.Endpoint,
		ApiEndpoint:        cluster.ApiEndpoint,
		EncryptionKeyId:    cluster.EncryptionKeyId,
		RestEndpoint:       cluster.RestEndpoint,
	}
}

func convertClusterToDescribeStructWithKAPI(cluster *schedv1.KafkaCluster) *describeStructWithKAPI {
	clusterStorage := strconv.Itoa(int(cluster.Storage))
	if clusterStorage == "-1" || cluster.InfiniteStorage {
		clusterStorage = "Infinite"
	}

	return &describeStructWithKAPI{
		Id:                 cluster.Id,
		Name:               cluster.Name,
		Type:               cluster.Deployment.Sku.String(),
		ClusterSize:        cluster.Cku,
		PendingClusterSize: cluster.PendingCku,
		NetworkIngress:     cluster.NetworkIngress,
		NetworkEgress:      cluster.NetworkEgress,
		Storage:            clusterStorage,
		ServiceProvider:    cluster.ServiceProvider,
		Region:             cluster.Region,
		Availability:       durabilityToAvaiablityNameMap[cluster.Durability.String()],
		Status:             cluster.Status.String(),
		Endpoint:           cluster.Endpoint,
		EncryptionKeyId:    cluster.EncryptionKeyId,
		RestEndpoint:       cluster.RestEndpoint,
		KAPI:               cluster.ApiEndpoint,
	}
}

func getKafkaClusterDescribeFields(cluster *schedv1.KafkaCluster, basicFields []string) []string {
	describeFields := basicFields
	if isDedicated(cluster) {
		describeFields = append(describeFields, "ClusterSize")
		if isExpanding(cluster) || isShrinking(cluster) {
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

func (c *clusterCommand) getUsageLimit(ctx context.Context, cku uint32) (int32, int32, error) {
	usageReply, err := c.Client.UsageLimits.GetUsageLimits(ctx)
	if err != nil || usageReply.UsageLimits == nil || len(usageReply.UsageLimits.GetCkuLimits()) == 0 || usageReply.UsageLimits.GetCkuLimits()[cku] == nil {
		return 0, 0, errors.Wrap(err, "Could not retrieve partition count usage limits. Please try again or contact support.")
	}
	partitionCount := usageReply.UsageLimits.GetCkuLimits()[cku].GetNumPartitions().GetValue()
	storageLimit := usageReply.UsageLimits.GetCkuLimits()[cku].Storage.GetValue()
	return partitionCount, storageLimit, nil
}

func isShrinking(cluster *schedv1.KafkaCluster) bool {
	return cluster.Status == schedv1.ClusterStatus_SHRINKING ||
		(cluster.PendingCku < cluster.Cku && cluster.PendingCku != 0)
}

func (c *clusterCommand) Cmd() *cobra.Command {
	return c.Command
}

func (c *clusterCommand) ServerComplete() []prompt.Suggest {
	var suggestions []prompt.Suggest
	clusters, err := pkafka.ListKafkaClusters(c.Client, c.EnvironmentId())
	if err != nil {
		return suggestions
	}
	for _, cluster := range clusters {
		suggestions = append(suggestions, prompt.Suggest{
			Text:        cluster.Id,
			Description: cluster.Name,
		})
	}
	return suggestions
}

func (c *clusterCommand) ServerCompletableChildren() []*cobra.Command {
	return c.completableChildren
}
